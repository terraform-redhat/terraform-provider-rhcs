package cms

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"net/http"

	"github.com/Masterminds/semver"
	. "github.com/onsi/gomega"
	client "github.com/openshift-online/ocm-sdk-go"
	v1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

// ImageVersion
type ImageVersion struct {
	ID           string `json:"id,omitempty"`
	RawID        string `json:"raw_id,omitempty"`
	ChannelGroup string `json:"channel_group,omitempty"`
	Enabled      bool   `json:"enabled,omitempty"`
	Default      bool   `json:"default,omitempty"`
	RosaEnabled  bool   `json:"rosa_enabled,omitempty"`
}

func EnabledVersions(connection *client.Connection, channel string, throttleVersion string, onlyRosa bool, upgradeAvailable ...bool) []*ImageVersion {
	search := "enabled= 't'"
	if channel != "" {
		search = fmt.Sprintf("%s and channel_group='%s'", search, channel)
	}
	if throttleVersion != "" {
		throttleVersion = `%` + throttleVersion + `%`
		search = fmt.Sprintf("%s and id like '%s'", search, throttleVersion)
	}
	if onlyRosa {
		search = fmt.Sprintf("%s and rosa_enabled = 't'", search)
	}
	if len(upgradeAvailable) == 1 && upgradeAvailable[0] {
		search = fmt.Sprintf("%s and available_upgrades != ''", search)
	}

	params := map[string]interface{}{
		"search": search,
		"size":   -1,
	}

	resp, err := ListVersions(connection, params)
	Expect(err).ToNot(HaveOccurred())
	Expect(resp.Status()).To(Equal(http.StatusOK))

	var imageVersionList []*ImageVersion
	versionItems := resp.Items().Slice()
	for _, version := range versionItems {
		imageVersion := &ImageVersion{
			ID:           version.ID(),
			RawID:        version.RawID(),
			ChannelGroup: version.ChannelGroup(),
			Enabled:      version.Enabled(),
			Default:      version.Default(),
			RosaEnabled:  version.ROSAEnabled(),
		}

		imageVersionList = append(imageVersionList, imageVersion)
	}

	return imageVersionList
}

func HCPEnabledVersions(connection *client.Connection, channel string, upgradeAvailable ...bool) []*ImageVersion {
	search := "enabled = 't' and hosted_control_plane_enabled='t' and rosa_enabled='t'"
	if channel != "" {
		search = fmt.Sprintf("%s and channel_group is '%s' ", search, channel)
	}
	if len(upgradeAvailable) == 1 && upgradeAvailable[0] {
		search = fmt.Sprintf("%s and available_upgrades != ''", search)
	}

	params := map[string]interface{}{
		"search": search,
		"size":   -1,
	}
	resp, err := ListVersions(connection, params)
	Expect(err).ToNot(HaveOccurred())
	Expect(resp.Status()).To(Equal(http.StatusOK))

	var imageVersionList []*ImageVersion
	versionItems := resp.Items().Slice()
	for _, version := range versionItems {
		imageVersion := &ImageVersion{
			ID:           version.ID(),
			RawID:        version.RawID(),
			ChannelGroup: version.ChannelGroup(),
			Enabled:      version.Enabled(),
			Default:      version.Default(),
			RosaEnabled:  version.ROSAEnabled(),
		}

		imageVersionList = append(imageVersionList, imageVersion)
	}

	return imageVersionList
}

// SortVersions sort the version list from lower to higher.
func SortVersions(versionList []*ImageVersion) []*ImageVersion {
	versionListIndexMap := make(map[string]*ImageVersion)
	var semverVersionList []*semver.Version
	for _, version := range versionList {
		index := fmt.Sprintf("%s-%s", version.RawID, version.ChannelGroup)
		versionListIndexMap[index] = version
		semverVersion, err := semver.NewVersion(index)
		if err != nil {
			panic(err)
		}
		semverVersionList = append(semverVersionList, semverVersion)
	}

	sort.Sort(semver.Collection(semverVersionList))
	var sortedImageVersionList []*ImageVersion
	for _, semverVersion := range semverVersionList {
		sortedImageVersionList = append(sortedImageVersionList, versionListIndexMap[semverVersion.Original()])
	}

	return sortedImageVersionList
}

func SortRawVersions(versionList []string) []string {
	sortedVersion := []string{}
	var semverVersionList []*semver.Version
	for _, version := range versionList {
		semverVersion, err := semver.NewVersion(version)
		if err != nil {
			panic(err)
		}
		semverVersionList = append(semverVersionList, semverVersion)
	}

	sort.Sort(semver.Collection(semverVersionList))
	for _, version := range semverVersionList {
		sortedVersion = append(sortedVersion, version.String())
	}
	return sortedVersion
}

// GetOneSpecifiedVersion returns a version with the specified index. The supported index string are one of
// ['latest', 'mid', 'oldest'], if the index string is an empty string or not belonged to the above list, the index will be
// a random value. If the version list is empty, will return nil directly.
func GetOneSpecifiedVersion(versionList []*ImageVersion, index string) *ImageVersion {
	length := len(versionList)
	if length == 0 {
		return nil
	}

	var version *ImageVersion
	switch index {
	case "latest":
		version = versionList[length-1]
	case "mid":
		version = versionList[length/2]
	case "oldest":
		version = versionList[0]
	default:
		randomIndex := h.NewRand().Intn(length)
		version = versionList[randomIndex]
	}

	return version
}

func FindAnUpgradeVersion(connection *client.Connection) string {
	timeNow := time.Now().UTC().Format(time.RFC3339)
	filterParam := map[string]interface{}{
		"search": fmt.Sprintf("enabled='t' and available_upgrades != '' and (end_of_life_timestamp > '%s' or end_of_life_timestamp is null)", timeNow),
	}
	resp, err := ListVersions(connection, filterParam)
	Expect(err).ToNot(HaveOccurred())
	Expect(resp.Status()).To(Equal(http.StatusOK))
	versionItems := resp.Items().Slice()
	randNum := h.NewRand().Intn(len(versionItems))
	return versionItems[randNum].ID()
}

// GetGreaterVersions will return a version list which is euqal or greater than the version provided as throttleVersion
func GetGreaterVersions(connection *client.Connection, throttleVersion string, channel string, onlyRosa bool, upgradeRequired bool) (versions []string) {
	versionIns := EnabledVersions(connection, channel, throttleVersion, onlyRosa, upgradeRequired)
	throttleVersionSem, err := semver.NewVersion(throttleVersion)
	Expect(err).ToNot(HaveOccurred())
	for _, version := range versionIns {
		currentVersion, err := semver.NewVersion(version.RawID)
		Expect(err).ToNot(HaveOccurred())
		if throttleVersionSem.LessThan(currentVersion) {
			versions = append(versions, version.ID)
		}
	}
	return
}
func GetGreaterOrEqualVersions(connection *client.Connection, throttleVersion string, channel string, onlyRosa bool, upgradeRequired bool) (versions []string) {
	versionIns := EnabledVersions(connection, channel, throttleVersion, onlyRosa, upgradeRequired)
	throttleVersionSem, err := semver.NewVersion(throttleVersion)
	Expect(err).ToNot(HaveOccurred())
	for _, version := range versionIns {
		fmt.Println(version.ID)
		currentVersion, err := semver.NewVersion(version.RawID)
		Expect(err).ToNot(HaveOccurred())
		if throttleVersionSem.LessThan(currentVersion) || throttleVersionSem.Equal(currentVersion) {
			versions = append(versions, version.RawID)
		}
	}
	return
}

func GetLowerVersions(connection *client.Connection, throttleVersion string, channel string, onlyRosa bool, upgradeRequired bool) (versions []string) {
	versionIns := EnabledVersions(connection, channel, throttleVersion, onlyRosa, upgradeRequired)
	throttleVersionSem, err := semver.NewVersion(throttleVersion)
	Expect(err).ToNot(HaveOccurred())
	for _, version := range versionIns {
		currentVersion, err := semver.NewVersion(version.RawID)
		Expect(err).ToNot(HaveOccurred())
		if currentVersion.LessThan(throttleVersionSem) {
			versions = append(versions, version.ID)
		}
	}
	return
}

func GetLowerOrEqualVersions(connection *client.Connection, throttleVersion string, channel string, onlyRosa bool, upgradeRequired bool) (versions []string) {
	versionIns := EnabledVersions(connection, channel, throttleVersion, onlyRosa, upgradeRequired)
	throttleVersionSem, err := semver.NewVersion(throttleVersion)
	Expect(err).ToNot(HaveOccurred())
	for _, version := range versionIns {
		currentVersion, err := semver.NewVersion(version.RawID)
		Expect(err).ToNot(HaveOccurred())
		if currentVersion.LessThan(throttleVersionSem) || currentVersion.Equal(throttleVersionSem) {
			versions = append(versions, version.RawID)
		}
	}
	return
}

// GetGreaterVersions will return a version list which is euqal or greater than the version provided as throttleVersion
func GetDefaultVersion(connection *client.Connection) *v1.Version {
	resp, err := ListVersions(connection, map[string]interface{}{"search": "default='true'"})
	Expect(err).ToNot(HaveOccurred())
	Expect(resp.Status()).To(Equal(http.StatusOK))
	version := resp.Items().Slice()[0]
	return version

}

// It will return all the versions lower that throttle version for the specified channel
func GetHcpLowerVersions(connection *client.Connection, throttleVersion string, channel string) (versions []string) {
	resp, _ := connection.ClustersMgmt().V1().Versions().List().Parameter("size", "-1").Send()
	throttleVersionSem, semVersionError := semver.NewVersion(throttleVersion)
	semver.NewVersion(throttleVersion)
	for _, ver := range resp.Items().Slice() {
		if semVersionError != nil {
			continue
		}
		if (ver.ChannelGroup() == channel) && ver.HostedControlPlaneEnabled() && ver.Enabled() {
			versionSem, _ := semver.NewVersion(ver.RawID())
			if versionSem.LessThan(throttleVersionSem) {
				versions = append(versions, ver.RawID())
			}
		}
	}
	return versions
}

// It will return all the versions higher that throttle version for the specified channel
func GetHcpHigherVersions(connection *client.Connection, throttleVersion string, channel string) (versions []string) {
	resp, _ := connection.ClustersMgmt().V1().Versions().List().Parameter("size", "-1").Send()
	throttleVersionSem, semVersionError := semver.NewVersion(throttleVersion)
	semver.NewVersion(throttleVersion)
	for _, ver := range resp.Items().Slice() {
		if semVersionError != nil {
			continue
		}
		if (ver.ChannelGroup() == channel) && ver.HostedControlPlaneEnabled() && ver.Enabled() {
			versionSem, _ := semver.NewVersion(ver.RawID())
			if versionSem.GreaterThan(throttleVersionSem) {
				versions = append(versions, ver.RawID())
			}
		}
	}
	return versions
}

// checks if upgradeVersion is a 'stream' upgrade for version
func IsStreamUpgrade(version string, upgradeVersion string, stream string) (isStreamUpgrade bool, err error) {
	if stream == CON.Y || stream == CON.Z || stream == CON.X {
		semVersion, semVersionError := semver.NewVersion(version)
		fmt.Printf("Testing %s and %s\n", version, upgradeVersion)
		if semVersionError == nil {
			semUpgradeVersion, semVersionError := semver.NewVersion(upgradeVersion)
			fmt.Printf("Testing %s and %s\n", semVersion.String(), semUpgradeVersion.String())
			if semVersionError == nil {
				if semVersion.Major() == semUpgradeVersion.Major() && semVersion.Minor() == semUpgradeVersion.Minor() && semVersion.Patch() < semUpgradeVersion.Patch() && stream == CON.Z {
					fmt.Printf("This version is z upgrade: %s\n", semUpgradeVersion.String())
					return true, nil
				} else if semVersion.Major() == semUpgradeVersion.Major() && semVersion.Minor() < semUpgradeVersion.Minor() && stream == CON.Y {
					fmt.Printf("This version is y upgrade: %s\n", semUpgradeVersion.String())
					return true, nil
				} else if semVersion.Major() < semUpgradeVersion.Major() && stream == CON.X {
					fmt.Printf("This version is x upgrade: %s\n", semUpgradeVersion.String())
					return true, nil
				} else {
					return false, nil
				}
			} else {
				err = fmt.Errorf("the version %s is invalid", upgradeVersion)
			}
		} else {
			err = fmt.Errorf("the version %s is invalid", version)
		}
	}
	return isStreamUpgrade, err
}

// It will return a list of versions which have available upgrades in the specified stream (x,y,z)
// channel mean channel group you are going to test
// stream means minor or patch upgrade like 4.x.y if you want x upgrade set stream=x, y upgrade set to stream=y
// step is a placeholder for future implementation. Only 1 supported now
func GetVersionsWithUpgrades(connection *client.Connection, channel string, stream string, rosaEnabled bool, hcpEnableRequired bool, step int) (imageVersionList []*ImageVersion, err error) {
	if step != 1 {
		return nil, fmt.Errorf("only 1 step support right now")
	}
	filters := []string{
		"enabled='t'",
		fmt.Sprintf("channel_group='%s'", channel),
		"available_upgrades != ''",
	}
	if rosaEnabled {
		filters = append(filters, "rosa_enabled='t'")
	}
	if hcpEnableRequired {
		filters = append(filters, "hosted_control_plane_enabled='t'")
	}

	filterParam := map[string]interface{}{
		"search": strings.Join(filters, " and "),
		"size":   "-1",
	}

	resp, err := ListVersions(connection, filterParam)
	Expect(err).ToNot(HaveOccurred())

	for _, ver := range resp.Items().Slice() {

		semVersion, semVersionError := semver.NewVersion(ver.RawID())
		if semVersionError != nil {
			continue
		}
		for _, avUpgrade := range ver.AvailableUpgrades() {
			semHigherVersion, _ := semver.NewVersion(avUpgrade)
			gocha := false
			switch stream {
			case CON.Z:
				if semVersion.Minor() == semHigherVersion.Minor() {
					imageVersion := &ImageVersion{
						ID:           ver.ID(),
						RawID:        ver.RawID(),
						ChannelGroup: ver.ChannelGroup(),
						Enabled:      ver.Enabled(),
						Default:      ver.Default(),
						RosaEnabled:  ver.ROSAEnabled(),
					}

					imageVersionList = append(imageVersionList, imageVersion)
					gocha = true
				}
			case CON.Y:
				if semVersion.Minor()+1 == semHigherVersion.Minor() {
					imageVersion := &ImageVersion{
						ID:           ver.ID(),
						RawID:        ver.RawID(),
						ChannelGroup: ver.ChannelGroup(),
						Enabled:      ver.Enabled(),
						Default:      ver.Default(),
						RosaEnabled:  ver.ROSAEnabled(),
					}

					imageVersionList = append(imageVersionList, imageVersion)
					gocha = true
				}
			default:
				return nil, fmt.Errorf("only y or z is allowed")
			}
			if gocha {
				break
			}
		}
	}
	imageVersionList = SortVersions(imageVersionList)
	return imageVersionList, err
}

// It will return a list of versions which have available upgrades in both y and z Streams
func GetVersionUpgradeTarget(orginalVersion string, stream string, availableUpgrades []string) (targetV string, err error) {
	semVersion, semVersionError := semver.NewVersion(orginalVersion)
	if semVersionError != nil {
		return "", err
	}
	for _, avUpgrade := range availableUpgrades {
		semHigherVersion, _ := semver.NewVersion(avUpgrade)
		switch stream {
		case CON.Z:
			if semVersion.Minor() == semHigherVersion.Minor() {
				targetV = avUpgrade
				return
			}
		case CON.Y:
			if semVersion.Minor()+1 == semHigherVersion.Minor() {
				targetV = avUpgrade
				return
			}
		default:
			return "", fmt.Errorf("only y or z is allowed")
		}

	}
	return
}

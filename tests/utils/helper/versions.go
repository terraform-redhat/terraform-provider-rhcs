package helper

***REMOVED***
***REMOVED***
	"sort"
	"time"

***REMOVED***

	"github.com/Masterminds/semver"
***REMOVED***
	client "github.com/openshift-online/ocm-sdk-go"
	v1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
***REMOVED***

// ImageVersion
type ImageVersion struct {
	ID           string `json:"id,omitempty"`
	RawID        string `json:"raw_id,omitempty"`
	ChannelGroup string `json:"channel_group,omitempty"`
	Enabled      bool   `json:"enabled,omitempty"`
	Default      bool   `json:"default,omitempty"`
	RosaEnabled  bool   `json:"rosa_enabled,omitempty"`
}

func ListVersions(connection *client.Connection, parameter ...map[string]interface{}***REMOVED*** (resp *v1.VersionsListResponse, err error***REMOVED*** {
	request := connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Versions(***REMOVED***.List(***REMOVED***
	if len(parameter***REMOVED*** == 1 {
		for k, v := range parameter[0] {
			request = request.Parameter(k, v***REMOVED***
***REMOVED***
	}
	resp, err = request.Send(***REMOVED***
	return
}

func EnabledVersions(connection *client.Connection, channel string, throttleVersion string, onlyRosa bool, upgradeAvailable ...bool***REMOVED*** []*ImageVersion {
	search := "enabled= 't'"
	if channel != "" {
		search = fmt.Sprintf("%s and channel_group='%s'", search, channel***REMOVED***
	}
	if throttleVersion != "" {
		throttleVersion = `%` + throttleVersion + `%`
		search = fmt.Sprintf("%s and id like '%s'", search, throttleVersion***REMOVED***
	}
	if onlyRosa {
		search = fmt.Sprintf("%s and rosa_enabled = 't'", search***REMOVED***
	}
	if len(upgradeAvailable***REMOVED*** == 1 && upgradeAvailable[0] {
		search = fmt.Sprintf("%s and available_upgrades != ''", search***REMOVED***
	}

	params := map[string]interface{}{
		"search": search,
		"size":   -1,
	}

	resp, err := ListVersions(connection, params***REMOVED***
	Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
	Expect(resp.Status(***REMOVED******REMOVED***.To(Equal(http.StatusOK***REMOVED******REMOVED***

	var imageVersionList []*ImageVersion
	versionItems := resp.Items(***REMOVED***.Slice(***REMOVED***
	for _, version := range versionItems {
		imageVersion := &ImageVersion{
			ID:           version.ID(***REMOVED***,
			RawID:        version.RawID(***REMOVED***,
			ChannelGroup: version.ChannelGroup(***REMOVED***,
			Enabled:      version.Enabled(***REMOVED***,
			Default:      version.Default(***REMOVED***,
			RosaEnabled:  version.ROSAEnabled(***REMOVED***,
***REMOVED***

		imageVersionList = append(imageVersionList, imageVersion***REMOVED***
	}

	return imageVersionList
}

func HCPEnabledVersions(connection *client.Connection, channel string, upgradeAvailable ...bool***REMOVED*** []*ImageVersion {
	// currentTime := time.Now(***REMOVED***.UTC(***REMOVED***.Format(time.RFC3339***REMOVED***
	search := "enabled = 't' and hosted_control_plane_enabled='t' and rosa_enabled='t'"
	if channel != "" {
		search = fmt.Sprintf("%s and channel_group is '%s' ", search, channel***REMOVED***
	}
	// search = search + fmt.Sprintf(" and (end_of_life_timestamp >'%s' or end_of_life_timestamp is null***REMOVED***", currentTime***REMOVED***
	if len(upgradeAvailable***REMOVED*** == 1 && upgradeAvailable[0] {
		search = fmt.Sprintf("%s and available_upgrades != ''", search***REMOVED***
	}

	params := map[string]interface{}{
		"search": search,
		"size":   -1,
	}
	resp, err := ListVersions(connection, params***REMOVED***
	Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
	Expect(resp.Status(***REMOVED******REMOVED***.To(Equal(http.StatusOK***REMOVED******REMOVED***

	var imageVersionList []*ImageVersion
	versionItems := resp.Items(***REMOVED***.Slice(***REMOVED***
	for _, version := range versionItems {
		imageVersion := &ImageVersion{
			ID:           version.ID(***REMOVED***,
			RawID:        version.RawID(***REMOVED***,
			ChannelGroup: version.ChannelGroup(***REMOVED***,
			Enabled:      version.Enabled(***REMOVED***,
			Default:      version.Default(***REMOVED***,
			RosaEnabled:  version.ROSAEnabled(***REMOVED***,
***REMOVED***

		imageVersionList = append(imageVersionList, imageVersion***REMOVED***
	}

	return imageVersionList
}

// SortVersions sort the version list from lower to higher.
func SortVersions(versionList []*ImageVersion***REMOVED*** []*ImageVersion {
	versionListIndexMap := make(map[string]*ImageVersion***REMOVED***
	var semverVersionList []*semver.Version
	for _, version := range versionList {
		index := fmt.Sprintf("%s-%s", version.RawID, version.ChannelGroup***REMOVED***
		versionListIndexMap[index] = version
		semverVersion, err := semver.NewVersion(index***REMOVED***
		if err != nil {
			panic(err***REMOVED***
***REMOVED***
		semverVersionList = append(semverVersionList, semverVersion***REMOVED***
	}

	sort.Sort(semver.Collection(semverVersionList***REMOVED******REMOVED***
	var sortedImageVersionList []*ImageVersion
	for _, semverVersion := range semverVersionList {
		sortedImageVersionList = append(sortedImageVersionList, versionListIndexMap[semverVersion.Original(***REMOVED***]***REMOVED***
	}

	return sortedImageVersionList
}

func SortRawVersions(versionList []string***REMOVED*** []string {
	sortedVersion := []string{}
	var semverVersionList []*semver.Version
	for _, version := range versionList {
		semverVersion, err := semver.NewVersion(version***REMOVED***
		if err != nil {
			panic(err***REMOVED***
***REMOVED***
		semverVersionList = append(semverVersionList, semverVersion***REMOVED***
	}

	sort.Sort(semver.Collection(semverVersionList***REMOVED******REMOVED***
	for _, version := range semverVersionList {
		sortedVersion = append(sortedVersion, version.String(***REMOVED******REMOVED***
	}
	return sortedVersion
}

// GetOneSpecifiedVersion returns a version with the specified index. The supported index string are one of
// ['latest', 'mid', 'oldest'], if the index string is an empty string or not belonged to the above list, the index will be
// a random value. If the version list is empty, will return nil directly.
func GetOneSpecifiedVersion(versionList []*ImageVersion, index string***REMOVED*** *ImageVersion {
	length := len(versionList***REMOVED***
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
		randomIndex := NewRand(***REMOVED***.Intn(length***REMOVED***
		version = versionList[randomIndex]
	}

	return version
}

func FindAnUpgradeVersion(connection *client.Connection***REMOVED*** string {
	timeNow := time.Now(***REMOVED***.UTC(***REMOVED***.Format(time.RFC3339***REMOVED***
	filterParam := map[string]interface{}{
		"search": fmt.Sprintf("enabled='t' and available_upgrades != '' and (end_of_life_timestamp > '%s' or end_of_life_timestamp is null***REMOVED***", timeNow***REMOVED***,
	}
	resp, err := ListVersions(connection, filterParam***REMOVED***
	Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
	Expect(resp.Status(***REMOVED******REMOVED***.To(Equal(http.StatusOK***REMOVED******REMOVED***
	versionItems := resp.Items(***REMOVED***.Slice(***REMOVED***
	randNum := NewRand(***REMOVED***.Intn(len(versionItems***REMOVED******REMOVED***
	return versionItems[randNum].ID(***REMOVED***
}

// GetGreaterVersions will return a version list which is euqal or greater than the version provided as throttleVersion
func GetGreaterVersions(connection *client.Connection, throttleVersion string, channel string, onlyRosa bool, upgradeRequired bool***REMOVED*** (versions []string***REMOVED*** {
	versionIns := EnabledVersions(connection, channel, throttleVersion, onlyRosa, upgradeRequired***REMOVED***
	throttleVersionSem, err := semver.NewVersion(throttleVersion***REMOVED***
	Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
	for _, version := range versionIns {
		currentVersion, err := semver.NewVersion(version.RawID***REMOVED***
		Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
		if throttleVersionSem.LessThan(currentVersion***REMOVED*** {
			versions = append(versions, version.ID***REMOVED***
***REMOVED***
	}
	return
}
func GetGreaterOrEqualVersions(connection *client.Connection, throttleVersion string, channel string, onlyRosa bool, upgradeRequired bool***REMOVED*** (versions []string***REMOVED*** {
	versionIns := EnabledVersions(connection, channel, throttleVersion, onlyRosa, upgradeRequired***REMOVED***
	throttleVersionSem, err := semver.NewVersion(throttleVersion***REMOVED***
	Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
	for _, version := range versionIns {
		fmt.Println(version.ID***REMOVED***
		currentVersion, err := semver.NewVersion(version.RawID***REMOVED***
		Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
		if throttleVersionSem.LessThan(currentVersion***REMOVED*** || throttleVersionSem.Equal(currentVersion***REMOVED*** {
			versions = append(versions, version.RawID***REMOVED***
***REMOVED***
	}
	return
}

func GetLowerVersions(connection *client.Connection, throttleVersion string, channel string, onlyRosa bool, upgradeRequired bool***REMOVED*** (versions []string***REMOVED*** {
	versionIns := EnabledVersions(connection, channel, throttleVersion, onlyRosa, upgradeRequired***REMOVED***
	throttleVersionSem, err := semver.NewVersion(throttleVersion***REMOVED***
	Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
	for _, version := range versionIns {
		currentVersion, err := semver.NewVersion(version.RawID***REMOVED***
		Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
		if currentVersion.LessThan(throttleVersionSem***REMOVED*** {
			versions = append(versions, version.ID***REMOVED***
***REMOVED***
	}
	return
}

func GetLowerOrEqualVersions(connection *client.Connection, throttleVersion string, channel string, onlyRosa bool, upgradeRequired bool***REMOVED*** (versions []string***REMOVED*** {
	versionIns := EnabledVersions(connection, channel, throttleVersion, onlyRosa, upgradeRequired***REMOVED***
	throttleVersionSem, err := semver.NewVersion(throttleVersion***REMOVED***
	Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
	for _, version := range versionIns {
		currentVersion, err := semver.NewVersion(version.RawID***REMOVED***
		Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
		if currentVersion.LessThan(throttleVersionSem***REMOVED*** || currentVersion.Equal(throttleVersionSem***REMOVED*** {
			versions = append(versions, version.RawID***REMOVED***
***REMOVED***
	}
	return
}

// GetGreaterVersions will return a version list which is euqal or greater than the version provided as throttleVersion
func GetDefaultVersion(connection *client.Connection***REMOVED*** *v1.Version {
	resp, err := ListVersions(connection, map[string]interface{}{"search": "default='true'"}***REMOVED***
	Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
	Expect(resp.Status(***REMOVED******REMOVED***.To(Equal(http.StatusOK***REMOVED******REMOVED***
	version := resp.Items(***REMOVED***.Slice(***REMOVED***[0]
	return version

}

// It will return all the versions lower that throttle version for the specified channel
func GetHcpLowerVersions(connection *client.Connection, throttleVersion string, channel string***REMOVED*** (versions []string***REMOVED*** {
	resp, _ := connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Versions(***REMOVED***.List(***REMOVED***.Parameter("size", "-1"***REMOVED***.Send(***REMOVED***
	throttleVersionSem, semVersionError := semver.NewVersion(throttleVersion***REMOVED***
	semver.NewVersion(throttleVersion***REMOVED***
	for _, ver := range resp.Items(***REMOVED***.Slice(***REMOVED*** {
		if semVersionError != nil {
			continue
***REMOVED***
		if (ver.ChannelGroup(***REMOVED*** == channel***REMOVED*** && ver.HostedControlPlaneEnabled(***REMOVED*** && ver.Enabled(***REMOVED*** {
			versionSem, _ := semver.NewVersion(ver.RawID(***REMOVED******REMOVED***
			if versionSem.LessThan(throttleVersionSem***REMOVED*** {
				versions = append(versions, ver.RawID(***REMOVED******REMOVED***
	***REMOVED***
***REMOVED***
	}
	return versions
}

// It will return all the versions higher that throttle version for the specified channel
func GetHcpHigherVersions(connection *client.Connection, throttleVersion string, channel string***REMOVED*** (versions []string***REMOVED*** {
	resp, _ := connection.ClustersMgmt(***REMOVED***.V1(***REMOVED***.Versions(***REMOVED***.List(***REMOVED***.Parameter("size", "-1"***REMOVED***.Send(***REMOVED***
	throttleVersionSem, semVersionError := semver.NewVersion(throttleVersion***REMOVED***
	semver.NewVersion(throttleVersion***REMOVED***
	for _, ver := range resp.Items(***REMOVED***.Slice(***REMOVED*** {
		if semVersionError != nil {
			continue
***REMOVED***
		if (ver.ChannelGroup(***REMOVED*** == channel***REMOVED*** && ver.HostedControlPlaneEnabled(***REMOVED*** && ver.Enabled(***REMOVED*** {
			versionSem, _ := semver.NewVersion(ver.RawID(***REMOVED******REMOVED***
			if versionSem.GreaterThan(throttleVersionSem***REMOVED*** {
				versions = append(versions, ver.RawID(***REMOVED******REMOVED***
	***REMOVED***
***REMOVED***
	}
	return versions
}

// checks if upgradeVersion is a 'stream' upgrade for version
func IsStreamUpgrade(version string, upgradeVersion string, stream string***REMOVED*** (isStreamUpgrade bool, err error***REMOVED*** {
	if stream == CON.Y || stream == CON.Z || stream == CON.X {
		semVersion, semVersionError := semver.NewVersion(version***REMOVED***
		fmt.Printf("Testing %s and %s\n", version, upgradeVersion***REMOVED***
		if semVersionError == nil {
			semUpgradeVersion, semVersionError := semver.NewVersion(upgradeVersion***REMOVED***
			fmt.Printf("Testing %s and %s\n", semVersion.String(***REMOVED***, semUpgradeVersion.String(***REMOVED******REMOVED***
			if semVersionError == nil {
				if semVersion.Major(***REMOVED*** == semUpgradeVersion.Major(***REMOVED*** && semVersion.Minor(***REMOVED*** == semUpgradeVersion.Minor(***REMOVED*** && semVersion.Patch(***REMOVED*** < semUpgradeVersion.Patch(***REMOVED*** && stream == CON.Z {
					fmt.Printf("This version is z upgrade: %s\n", semUpgradeVersion.String(***REMOVED******REMOVED***
					return true, nil
		***REMOVED*** else if semVersion.Major(***REMOVED*** == semUpgradeVersion.Major(***REMOVED*** && semVersion.Minor(***REMOVED*** < semUpgradeVersion.Minor(***REMOVED*** && stream == CON.Y {
					fmt.Printf("This version is y upgrade: %s\n", semUpgradeVersion.String(***REMOVED******REMOVED***
					return true, nil
		***REMOVED*** else if semVersion.Major(***REMOVED*** < semUpgradeVersion.Major(***REMOVED*** && stream == CON.X {
					fmt.Printf("This version is x upgrade: %s\n", semUpgradeVersion.String(***REMOVED******REMOVED***
					return true, nil
		***REMOVED*** else {
					return false, nil
		***REMOVED***
	***REMOVED*** else {
				err = fmt.Errorf("the version %s is invalid", upgradeVersion***REMOVED***
	***REMOVED***
***REMOVED*** else {
			err = fmt.Errorf("the version %s is invalid", version***REMOVED***
***REMOVED***
	}
	return isStreamUpgrade, err
}

// It will return a list of versions which have available upgrades in the specified stream (x,y,z***REMOVED***
func GetVersionsWithUpgrades(connection *client.Connection, channel string, stream string, hcpEnabled bool***REMOVED*** (imageVersionList []*ImageVersion, err error***REMOVED*** {
	timeNow := time.Now(***REMOVED***.UTC(***REMOVED***.Format(time.RFC3339***REMOVED***
	var enabledString = ""
	if hcpEnabled {
		enabledString = "t"
	} else {
		enabledString = "f"
	}
	filterParam := map[string]interface{}{
		"search": fmt.Sprintf("enabled='t' and hosted_control_plane_enabled='%s' and rosa_enabled='t' and channel_group='%s' and available_upgrades != '' and (end_of_life_timestamp > '%s' or end_of_life_timestamp is null***REMOVED***", enabledString, channel, timeNow***REMOVED***,
		"size":   "-1",
	}

	resp, err := ListVersions(connection, filterParam***REMOVED***
	Expect(err***REMOVED***.ToNot(HaveOccurred(***REMOVED******REMOVED***
	Expect(resp.Status(***REMOVED******REMOVED***.To(Equal(http.StatusOK***REMOVED******REMOVED***

	for _, ver := range resp.Items(***REMOVED***.Slice(***REMOVED*** {

		semVersion, semVersionError := semver.NewVersion(ver.RawID(***REMOVED******REMOVED***
		if semVersionError != nil {
			continue
***REMOVED***

		var upgrade semver.Version
		switch stream {
		case CON.X:
			upgrade = semVersion.IncMajor(***REMOVED***
		case CON.Y:
			semVersion.IncMinor(***REMOVED***
		case CON.Z:
			upgrade = semVersion.IncPatch(***REMOVED***
		default:
			err = fmt.Errorf("the stream %s is invalid", stream***REMOVED***
***REMOVED***

		for _, avUpgrade := range ver.AvailableUpgrades(***REMOVED*** {
			semUpgrade, _ := semver.NewVersion(avUpgrade***REMOVED***
			semUpgradeNoPreRelease, _ := semUpgrade.SetPrerelease(""***REMOVED*** // removing pre-release from version string
			// fmt.Printf("avUpgrade=%s, reference upgrade = %s\n", semUpgradeNoPreRelease, &upgrade***REMOVED***
			if semUpgradeNoPreRelease.GreaterThan(&upgrade***REMOVED*** || semUpgradeNoPreRelease.Equal(&upgrade***REMOVED*** {

				zStreamAndIsZStreamUpgrade := stream == CON.Z && semUpgradeNoPreRelease.Minor(***REMOVED*** == semVersion.Minor(***REMOVED*** && semUpgradeNoPreRelease.Major(***REMOVED*** == semVersion.Major(***REMOVED***
				yStreamAndIsYStreamUpgrade := stream == CON.Y && semUpgradeNoPreRelease.Major(***REMOVED*** == semVersion.Major(***REMOVED***
				xStreamAndIsXStreamUpgrade := stream == CON.X

				if zStreamAndIsZStreamUpgrade || yStreamAndIsYStreamUpgrade || xStreamAndIsXStreamUpgrade {
					imageVersion := &ImageVersion{
						ID:           ver.ID(***REMOVED***,
						RawID:        ver.RawID(***REMOVED***,
						ChannelGroup: ver.ChannelGroup(***REMOVED***,
						Enabled:      ver.Enabled(***REMOVED***,
						Default:      ver.Default(***REMOVED***,
						RosaEnabled:  ver.ROSAEnabled(***REMOVED***,
			***REMOVED***

					imageVersionList = append(imageVersionList, imageVersion***REMOVED***
					break
		***REMOVED***
	***REMOVED***
***REMOVED***
	}
	imageVersionList = SortVersions(imageVersionList***REMOVED***
	return imageVersionList, err
}

// It will return a list of versions which have available upgrades in both y and z Streams
func GetVersionsWithZYUpgrades(connection *client.Connection, channel string, hcpEnabled bool***REMOVED*** (imageVersionList *ImageVersion, err error***REMOVED*** {
	lowVersionWithZUpgList, _ := GetVersionsWithUpgrades(connection, channel, CON.Z, true***REMOVED***
	lowVersionWithYUpgList, _ := GetVersionsWithUpgrades(connection, channel, CON.Y, true***REMOVED***
	var versionWithXYUpgrade *ImageVersion

	if len(lowVersionWithZUpgList***REMOVED*** == 0 || len(lowVersionWithYUpgList***REMOVED*** == 0 {
		err = fmt.Errorf("No lower version was found"***REMOVED***
	} else {
		for index := len(lowVersionWithZUpgList***REMOVED*** - 1; index >= 0; index-- {
			versionWithZUpg := lowVersionWithZUpgList[index]
			for yIndex := len(lowVersionWithYUpgList***REMOVED*** - 1; yIndex >= 0; yIndex-- {
				versionWithYUpg := lowVersionWithYUpgList[yIndex]
				if versionWithZUpg.RawID == versionWithYUpg.RawID {
					versionWithXYUpgrade = versionWithYUpg
					fmt.Sprintf("Found version: %s", versionWithYUpg.RawID***REMOVED***
					break
		***REMOVED***
				if versionWithXYUpgrade != nil {
					break
		***REMOVED***
	***REMOVED***
***REMOVED***
	}

	return versionWithXYUpgrade, err
}

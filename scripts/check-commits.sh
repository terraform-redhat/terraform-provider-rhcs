#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset

__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

remote=$(git remote -v | (grep "https://github\.com/terraform-redhat/terraform-provider-rhcs" || true) | head -n 1 | awk '{ print $1 }')
if [ -z $remote ]; then
    echo "could not find remote for github.com/terraform-redhat/terraform-provider-rhcs"
    exit 1
fi
master_branch="$remote/main"
echo "main branch ${master_branch}"

pull_request_number=$(jq --raw-output .pull_request.number "$GITHUB_EVENT_PATH")

commit_messages=$(curl -s \
  "https://api.github.com/repos/$GITHUB_REPOSITORY/pulls/$pull_request_number/commits" | \
  jq -r '.[].commit.message')

for commit in ${commit_messages};
do
    echo "commit : ${commit}"
    ${__dir}/check-commit-message.sh "${commit}"
done


exit 0
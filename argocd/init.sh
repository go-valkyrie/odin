#!/bin/bash
set -e
set -o pipefail
cleanup_values=false
values_file=$(echo "$ARGOCD_APP_PARAMETERS" |  jq -r '.[] | select(.name=="valuesFile") | .string')
if [[ "$values_file" = "" ]]; then
  values=$(echo $ARGOCD_APP_PARAMETERS | jq '.[] | select(.name=="values") | if .map != null then .map else {} end')
  values_file=$(mktemp -t values.XXXX)
  cleanup_values=true
  echo "$values" >> $values_file
else
  echo "Using values file: $values_file"
fi

stat "$values_file" > /dev/null

odin template -t json -f "$values_file"

if $cleanup_values; then
  rm $values_file
fi
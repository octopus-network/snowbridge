#!/usr/bin/env bash
set -eu

source scripts/set-env.sh
source scripts/xcm-helper.sh

enable_gateway() {
    local transact_call="0x330400"
    send_governance_transact_from_relaychain $BRIDGE_HUB_PARAID "$transact_call"
}

disable_gateway() {
    local transact_call="0x330401"
    send_governance_transact_from_relaychain $BRIDGE_HUB_PARAID "$transact_call"
}

read -p "Enable gateway? (Y/N): " confirm

if [[ $confirm == [yY] || $confirm == [yY][eE][sS] ]]; then
    enable_gateway
else
    disable_gateway
fi

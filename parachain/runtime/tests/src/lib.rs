// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2023 Snowfork <hello@snowfork.com>

#![cfg(test)]

mod test_cases;

use asset_hub_rococo_runtime::xcm_config::bridging::to_ethereum::DefaultBridgeHubEthereumBaseFee;
use bridge_hub_rococo_runtime::{xcm_config::XcmConfig, Runtime, RuntimeEvent, SessionKeys};
use codec::Decode;
use cumulus_primitives_core::XcmError::{FailedToTransactAsset, NotHoldingFees, Unroutable};
use parachains_common::{AccountId, AuraId};
use sp_core::H160;
use sp_keyring::AccountKeyring::Alice;

pub fn collator_session_keys() -> bridge_hub_test_utils::CollatorSessionKeys<Runtime> {
	bridge_hub_test_utils::CollatorSessionKeys::new(
		AccountId::from(Alice),
		AccountId::from(Alice),
		SessionKeys { aura: AuraId::from(Alice.public()) },
	)
}

#[test]
pub fn transfer_token_to_ethereum_works() {
	test_cases::send_transfer_token_message_success::<Runtime, XcmConfig>(
		collator_session_keys(),
		1013,
		1000,
		H160::random(),
		H160::random(),
		DefaultBridgeHubEthereumBaseFee::get(),
		Box::new(|runtime_event_encoded: Vec<u8>| {
			match RuntimeEvent::decode(&mut &runtime_event_encoded[..]) {
				Ok(RuntimeEvent::EthereumOutboundQueue(event)) => Some(event),
				_ => None,
			}
		}),
	)
}

#[test]
pub fn transfer_token_from_non_assethub_to_ethereum_unroutable() {
	test_cases::send_transfer_token_message_failure::<Runtime, XcmConfig>(
		collator_session_keys(),
		1013,
		1001,
		DefaultBridgeHubEthereumBaseFee::get() + 1_000_000_000,
		H160::random(),
		H160::random(),
		// fee not enough
		1_000_000_000,
		Unroutable,
	)
}

#[test]
pub fn unpaid_transfer_token_to_ethereum_fails_with_barrier() {
	test_cases::send_unpaid_transfer_token_message::<Runtime, XcmConfig>(
		collator_session_keys(),
		1013,
		1000,
		H160::random(),
		H160::random(),
	)
}

#[test]
pub fn transfer_token_to_ethereum_fee_not_enough() {
	test_cases::send_transfer_token_message_failure::<Runtime, XcmConfig>(
		collator_session_keys(),
		1013,
		1000,
		DefaultBridgeHubEthereumBaseFee::get() + 1_000_000_000,
		H160::random(),
		H160::random(),
		// fee not enough
		1_000_000_000,
		NotHoldingFees,
	)
}

#[test]
pub fn transfer_token_to_ethereum_insufficient_fund() {
	test_cases::send_transfer_token_message_failure::<Runtime, XcmConfig>(
		collator_session_keys(),
		1013,
		1000,
		1_000_000_000,
		H160::random(),
		H160::random(),
		DefaultBridgeHubEthereumBaseFee::get(),
		FailedToTransactAsset("InsufficientBalance"),
	)
}

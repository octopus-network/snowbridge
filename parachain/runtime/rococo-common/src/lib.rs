// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2023 Snowfork <hello@snowfork.com>
//! # Rococo Common
//!
//! Config used for the Rococo asset hub and bridge hub runtimes.
#![cfg_attr(not(feature = "std"), no_std)]

use frame_support::parameter_types;
use xcm::opaque::lts::NetworkId;

parameter_types! {
	// Network and location for the Ethereum chain.
	pub EthereumNetwork: NetworkId = NetworkId::Ethereum { chain_id: 15 };
}

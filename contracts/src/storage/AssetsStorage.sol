// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2023 Snowfork <hello@snowfork.com>
pragma solidity 0.8.23;

import {TokenInfo, ParaID} from "../Types.sol";

library AssetsStorage {
    struct Layout {
        // Token registry by token address
        mapping(address token => TokenInfo) tokenRegistry;
        address assetHubAgent;
        ParaID assetHubParaID;
        // XCM fee charged by AssetHub for registering a token (DOT)
        uint128 assetHubCreateAssetFee;
        // XCM fee charged by AssetHub for receiving a token from the Gateway (DOT)
        uint128 assetHubReserveTransferFee;
        // Extra fee for registering a token, to discourage spamming (Ether)
        uint256 registerTokenFee;
        // Token registry by tokenID
        mapping(bytes32 tokenID => TokenInfo) tokenRegistryByID;
    }

    bytes32 internal constant SLOT = keccak256("org.snowbridge.storage.assets");

    function layout() internal pure returns (Layout storage $) {
        bytes32 slot = SLOT;
        assembly {
            $.slot := slot
        }
    }
}

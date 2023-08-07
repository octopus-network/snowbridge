use hex_literal::hex;

pub const ETHEREUM_API: &str = "ws://localhost:8546";
pub const ETHEREUM_HTTP_API: &str = "http://localhost:8545";
pub const BRIDGE_HUB_WS_URL: &str = "ws://127.0.0.1:11144";
pub const BRIDGE_HUB_PARA_ID: u32 = 1013;

pub const ASSET_HUB_WS_URL: &str = "ws://127.0.0.1:12144";

pub const ASSET_HUB_WEIGHT_REQUIRED: u64 = 3000000000;
pub const ASSET_HUB_PROOF_SIZE_REQUIRED: u64 = 18000;
pub const BRIDGE_HUB_FEE_REQUIRED: u128 = 1000000000000;

pub const ETHEREUM_CHAIN_ID: u64 = 15;
pub const ETHEREUM_KEY: &str = "0x5e002a1af63fd31f1c25258f3082dc889762664cb8f218d86da85dff8b07b342";
pub const ETHEREUM_ADDRESS: [u8; 20] = hex!("90A987B944Cb1dCcE5564e5FDeCD7a54D3de27Fe");

// GatewayProxy in local setup
pub const GATEWAY_PROXY_CONTRACT: [u8; 20] = hex!("EDa338E4dC46038493b885327842fD3E301CaB39");

// Agent for sibling parachain
pub const SIBLING_AGENT_ID: [u8; 32] =
    hex!("a0a4510eb0ccc47db0cb6bdadf5db6641ddd423668c4b11fc51889d23794eb5d");
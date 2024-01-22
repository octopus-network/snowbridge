
import { channelStatusInfo, contextFactory, bridgeStatusInfo, planSendToken, doSendToken, trackSendToken } from '@snowbridge/api'
import { Wallet, ethers } from 'ethers'

const ETHEREUM_WS_API = 'ws://127.0.0.1:8546'
const RELAY_CHAIN_WS_URL = 'ws://127.0.0.1:9944'
const ASSET_HUB_WS_URL = 'ws://127.0.0.1:12144'
const BRIDGE_HUB_WS_URL = 'ws://127.0.0.1:11144'
const GATEWAY_CONTRACT = '0xEDa338E4dC46038493b885327842fD3E301CaB39'
const BEEFY_CONTRACT = '0x992B9df075935E522EC7950F37eC8557e86f6fdb'
const WETH_CONTRACT = '0x87d1f7fdfEe7f651FaBc8bFCB6E086C278b77A7d'
const ASSET_HUB_CHANNEL_ID = '0xc173fac324158e77fb5840738a1a541f633cbec8884c6a601c567d2b376a0539'
const PRIMARY_GOVERNANCE_CHANNEL_ID = '0x0000000000000000000000000000000000000000000000000000000000000001'
const SECONDARY_GOVERNANCE_CHANNEL_ID = '0x0000000000000000000000000000000000000000000000000000000000000002'

const monitor = async () => {
    const context = await contextFactory({
        ethereum: { url: ETHEREUM_WS_API },
        polkadot: {
            url: {
                bridgeHub: BRIDGE_HUB_WS_URL,
                assetHub: ASSET_HUB_WS_URL,
                relaychain: RELAY_CHAIN_WS_URL,
            },
        },
        appContracts: {
            gateway: GATEWAY_CONTRACT,
            beefy: BEEFY_CONTRACT,
        },
    })

    const status = await bridgeStatusInfo(context)
    console.log('Bridge Status:', status)
    const assethub = await channelStatusInfo(context, ASSET_HUB_CHANNEL_ID)
    console.log('Asset Hub Channel:', assethub)
    const primary_gov = await channelStatusInfo(context, PRIMARY_GOVERNANCE_CHANNEL_ID)
    console.log('Primary Governance Channel:', primary_gov)
    const secondary_gov = await channelStatusInfo(context, SECONDARY_GOVERNANCE_CHANNEL_ID)
    console.log('Secondary Governance Channel:', secondary_gov)

    const signer = new Wallet('0x5e002a1af63fd31f1c25258f3082dc889762664cb8f218d86da85dff8b07b342', context.ethereum.api)
    const plan = await planSendToken(context, signer, '5CiPPseXPECbkjWCa6MnjNokrgYjMqmKndv2rSnekmSK2DjL', WETH_CONTRACT, BigInt(1000))
    console.log('Plan:', plan)
    const result = await doSendToken(context, signer, plan)
    console.log('Execute:', result)
    await trackSendToken(context, result);
    await trackSendToken(context, result);
    // Get execution block and message id
    // Watch bridgehub beacon client until block is included, emit included in light client
    // Watch bridgehub nonce update for channel and message id and dispatched xcm
    // Watch assethub for messagequeue processed and foreign assets issued to owner with amount.
}


monitor()
    .then(() => process.exit(0))
    .catch((error) => {
        console.error(error)
        process.exit(1)
    })

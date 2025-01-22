package dex

import (
	"fmt"
)

func (pfe *fuzzDexExecutor) init(args *fuzzDexExecutorInitArgs) error {
	pfe.wrewaTokenId = args.wrewaTokenId
	pfe.moaTokenId = args.moaTokenId
	pfe.busdTokenId = args.busdTokenId
	pfe.wemeLpTokenId = args.wemeLpTokenId
	pfe.webuLpTokenId = args.webuLpTokenId
	pfe.wemeFarmTokenId = args.wemeFarmTokenId
	pfe.webuFarmTokenId = args.webuFarmTokenId
	pfe.moaFarmTokenId = args.moaFarmTokenId
	pfe.numUsers = args.numUsers
	pfe.numEvents = args.numEvents
	pfe.removeLiquidityProb = args.removeLiquidityProb
	pfe.addLiquidityProb = args.addLiquidityProb
	pfe.swapProb = args.swapProb
	pfe.queryPairsProb = args.queryPairsProb
	pfe.enterFarmProb = args.enterFarmProb
	pfe.exitFarmProb = args.exitFarmProb
	pfe.claimRewardsProb = args.claimRewardsProb
	pfe.increaseBlockNonceProb = args.increaseBlockNonceProb
	pfe.removeLiquidityMaxValue = args.removeLiquidityMaxValue
	pfe.addLiquidityMaxValue = args.addLiquidityMaxValue
	pfe.swapMaxValue = args.swapMaxValue
	pfe.enterFarmMaxValue = args.enterFarmMaxValue
	pfe.exitFarmMaxValue = args.exitFarmMaxValue
	pfe.claimRewardsMaxValue = args.claimRewardsMaxValue
	pfe.blockNonceIncrease = args.blockNonceIncrease
	pfe.farmers = make(map[int]FarmerInfo)
	pfe.currentFarmTokenNonce = make(map[string]int)

	pfe.world.Clear()

	pfe.ownerAddress = "address:fuzz_owner"
	pfe.wemeFarmAddress = "sc:fuzz_dex_weme_farm"
	pfe.webuFarmAddress = "sc:fuzz_dex_webu_farm"
	pfe.moaFarmAddress = "sc:fuzz_dex_moa_farm"
	pfe.wemeSwapAddress = "sc:fuzz_dex_weme_swap"
	pfe.webuSwapAddress = "sc:fuzz_dex_webu_swap"

	pfe.currentFarmTokenNonce[pfe.wemeFarmTokenId] = 0
	pfe.currentFarmTokenNonce[pfe.webuFarmTokenId] = 0
	pfe.currentFarmTokenNonce[pfe.moaFarmTokenId] = 0

	pfe.farms[0] = Farm{
		address:      pfe.wemeFarmAddress,
		farmToken:    pfe.wemeFarmTokenId,
		farmingToken: pfe.wemeLpTokenId,
		rewardToken:  pfe.moaTokenId,
	}
	pfe.farms[1] = Farm{
		address:      pfe.webuFarmAddress,
		farmToken:    pfe.webuFarmTokenId,
		farmingToken: pfe.webuLpTokenId,
		rewardToken:  pfe.moaTokenId,
	}
	pfe.farms[2] = Farm{
		address:      pfe.moaFarmAddress,
		farmToken:    pfe.moaFarmTokenId,
		farmingToken: pfe.moaTokenId,
		rewardToken:  pfe.moaTokenId,
	}

	pfe.swaps[0] = SwapPair{
		address:     pfe.wemeSwapAddress,
		lpToken:     pfe.wemeLpTokenId,
		firstToken:  pfe.wrewaTokenId,
		secondToken: pfe.moaTokenId,
	}
	pfe.swaps[1] = SwapPair{
		address:     pfe.webuSwapAddress,
		lpToken:     pfe.webuLpTokenId,
		firstToken:  pfe.wrewaTokenId,
		secondToken: pfe.busdTokenId,
	}

	// users
	dcdtString := pfe.fullOfDcdtWalletString()
	for i := 1; i <= args.numUsers; i++ {
		err := pfe.executeStep(fmt.Sprintf(`
		{
			"step": "setState",
			"accounts": {
				"%s": {
					"nonce": "0",
					"balance": "0",
					"storage": {},
					"dcdt": {
						%s
					},
					"code": ""
				}
			}
		}`,
			pfe.userAddress(i),
			dcdtString,
		))
		if err != nil {
			return err
		}
	}
	err := pfe.executeStep(fmt.Sprintf(`
	{
		"step": "setState",
		"accounts": {
			"%s": {
				"nonce": "0",
				"balance": "1,000,000,000,000,000,000,000,000,000,000",
				"storage": {},
				"code": ""
			}
		}
	}`,
		pfe.ownerAddress,
	))
	if err != nil {
		return err
	}

	// swaps
	err = pfe.setupPair(pfe.wemeSwapAddress, pfe.wrewaTokenId, pfe.moaTokenId, pfe.wemeLpTokenId, pfe.ownerAddress, pfe.ownerAddress)
	if err != nil {
		return err
	}

	err = pfe.setupPair(pfe.webuSwapAddress, pfe.wrewaTokenId, pfe.busdTokenId, pfe.webuLpTokenId, pfe.ownerAddress, pfe.ownerAddress)
	if err != nil {
		return err
	}

	// farms
	err = pfe.setupFarm(pfe.wemeFarmAddress, pfe.wemeFarmTokenId, pfe.wemeLpTokenId, pfe.moaTokenId, pfe.ownerAddress, pfe.ownerAddress)
	if err != nil {
		return err
	}

	err = pfe.setupFarm(pfe.webuFarmAddress, pfe.webuFarmTokenId, pfe.webuLpTokenId, pfe.moaTokenId, pfe.ownerAddress, pfe.ownerAddress)
	if err != nil {
		return err
	}

	err = pfe.setupFarm(pfe.moaFarmAddress, pfe.moaFarmTokenId, pfe.moaTokenId, pfe.moaTokenId, pfe.ownerAddress, pfe.ownerAddress)
	if err != nil {
		return err
	}

	// configs
	err = pfe.setFeeOn(pfe.wemeSwapAddress, pfe.wemeFarmAddress, pfe.moaTokenId, pfe.ownerAddress)
	if err != nil {
		return err
	}

	err = pfe.setFeeOn(pfe.webuSwapAddress, pfe.webuFarmAddress, pfe.moaTokenId, pfe.ownerAddress)
	if err != nil {
		return err
	}

	err = pfe.whitelist(pfe.ownerAddress, pfe.wemeSwapAddress, pfe.webuSwapAddress)
	if err != nil {
		return err
	}

	err = pfe.addTrustedSwapPair(pfe.ownerAddress, pfe.webuSwapAddress, pfe.wemeSwapAddress, pfe.wrewaTokenId, pfe.moaTokenId)
	if err != nil {
		return err
	}

	pfe.log("init ok")
	return nil
}

func (pfe *fuzzDexExecutor) setupPair(swapAddress, firstTokenId, secondTokenId, lpTokenId, routerAddress, owner string) error {
	return pfe.executeStep(fmt.Sprintf(`
		{
			"step": "setState",
			"accounts": {
				"%s": {
					"nonce": "0",
					"balance": "0",
					"dcdt": {
						"str:%s": {
							"roles": [
								"DCDTRoleLocalMint",
								"DCDTRoleLocalBurn"
							]
						}
					},
					"storage": {
						"str:first_token_id": "str:%s",
						"str:second_token_id": "str:%s",
						"str:state": "1",
						"str:lpTokenIdentifier": "str:%s",
						"str:router_address": "%s",
						"str:router_owner_address": "%s",
						"str:total_fee_percent": "300",
						"str:special_fee_percent": "050",
						"str:burn_tokens_gas_limit": "500000",
						"str:mint_tokens_gas_limit": "500000",
						"str:send_fee_gas_limit": "25,000,000",
						"str:extern_swap_gas_limit": "50,000,000"
					},
					"code": "file:dex_pair.wasm",
					"owner": "%s"
				}
			}
		}`,
		swapAddress,
		lpTokenId,
		firstTokenId,
		secondTokenId,
		lpTokenId,
		routerAddress,
		owner,
		owner,
	))
}

func (pfe *fuzzDexExecutor) setupFarm(farmAddress, farmTokenId, enterFarmTokenId, rewardTokenId, routerAddress, owner string) error {
	return pfe.executeStep(fmt.Sprintf(`
		{
			"step": "setState",
			"accounts": {
				"%s": {
					"nonce": "0",
					"balance": "0",
					"dcdt": {
						"str:%s": {
							"roles": [
								"DCDTRoleNFTCreate",
								"DCDTRoleNFTAddQuantity",
								"DCDTRoleNFTBurn"
							]
						},
						"str:%s": {
							"roles": [
								"DCDTRoleLocalMint",
								"DCDTRoleLocalBurn"
							]
						},
						"str:%s": {
							"roles": [
								"DCDTRoleLocalMint",
								"DCDTRoleLocalBurn"
							]
						}
					},
					"storage": {
						"str:farming_token_id": "str:%s",
						"str:farm_token_id": "str:%s",
						"str:reward_token_id": "str:%s",
						"str:router_address": "%s",
						"str:state": "1",
						"str:owner": "%s",
						"str:minimum_farming_epochs": "2",
						"str:burn_tokens_gas_limit": "5,000,000",
						"str:mint_tokens_gas_limit": "5,000,000",
						"str:locked_rewards_apr_multiplier": "2",
						"str:division_safety_constant": "1000000000000",
						"str:create_farm_tokens_gas_limit": "5000000",
						"str:produce_rewards_enabled": "1",
						"str:per_block_reward_amount": "10000000000000000",
						"str:penalty_percent": "0"
					},
					"code": "file:dex_farm.wasm",
					"owner": "%s"
				}
			}
		}`,
		farmAddress,
		farmTokenId,
		enterFarmTokenId,
		rewardTokenId,
		enterFarmTokenId,
		farmTokenId,
		rewardTokenId,
		routerAddress,
		owner,
		owner,
	))
}

func (pfe *fuzzDexExecutor) setFeeOn(swapAddress, farmAddress, feeToken, ownerAddress string) error {
	_, err := pfe.executeTxStep(fmt.Sprintf(`
		{
			"step": "scCall",
			"txId": "set-fee-on",
			"tx": {
				"from": "%s",
				"to": "%s",
				"value": "0",
				"function": "setFeeOn",
				"arguments": [
					"1",
					"%s",
					"str:%s"
				],
				"gasLimit": "10,000,000",
				"gasPrice": "0"
			},
			"expect": {
				"out": [],
				"status": "",
				"logs": [],
				"gas": "*",
				"refund": "*"
			}
		}`,
		ownerAddress,
		swapAddress,
		farmAddress,
		feeToken,
	))
	return err
}

func (pfe *fuzzDexExecutor) whitelist(ownerAddress, swapAddressToConfig, swapAddressToWhitelist string) error {
	_, err := pfe.executeTxStep(fmt.Sprintf(`
		{
			"step": "scCall",
			"txId": "whitelist",
			"tx": {
				"from": "%s",
				"to": "%s",
				"value": "0",
				"function": "whitelist",
				"arguments": [
					"%s"
				],
				"gasLimit": "10,000,000",
				"gasPrice": "0"
			},
			"expect": {
				"out": [],
				"status": "",
				"logs": [],
				"gas": "*",
				"refund": "*"
			}
		}`,
		ownerAddress,
		swapAddressToConfig,
		swapAddressToWhitelist,
	))
	return err
}

func (pfe *fuzzDexExecutor) addTrustedSwapPair(ownerAddress, swapAddressToConfig, swapAddressToAdd, firstTokenId, secondTokenId string) error {
	_, err := pfe.executeTxStep(fmt.Sprintf(`
		{
			"step": "scCall",
			"txId": "whitelist",
			"tx": {
				"from": "%s",
				"to": "%s",
				"value": "0",
				"function": "addTrustedSwapPair",
				"arguments": [
					"%s",
					"str:%s",
					"str:%s"
				],
				"gasLimit": "10,000,000",
				"gasPrice": "0"
			},
			"expect": {
				"out": [],
				"status": "",
				"logs": [],
				"gas": "*",
				"refund": "*"
			}
		}`,
		ownerAddress,
		swapAddressToConfig,
		swapAddressToAdd,
		firstTokenId,
		secondTokenId,
	))
	return err
}

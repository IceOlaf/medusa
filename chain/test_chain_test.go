package chain

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/stretchr/testify/assert"
	"github.com/trailofbits/medusa/compilation/platforms"
	"github.com/trailofbits/medusa/utils"
	"github.com/trailofbits/medusa/utils/testutils"
	"math/big"
	"math/rand"
	"testing"
)

// verifyChain verifies various state properties in a TestChain, such as if previous block hashes are correct,
// timestamps are in order, etc.
func verifyChain(t *testing.T, chain *TestChain) {
	// Assert there are blocks
	assert.Greater(t, len(chain.blocks), 0)

	// Assert that the head is the last block
	assert.EqualValues(t, chain.blocks[len(chain.blocks)-1], chain.Head())

	// Loop through all blocks
	// Note: We use the API here rather than internally committed blocks (chain.blocks) to validate spoofed blocks
	// from height jumps as well.
	for i := int(chain.HeadBlockNumber()); i >= 0; i-- {
		// Verify our count of messages, message results, and receipts match.
		currentBlock, err := chain.BlockFromNumber(uint64(i))
		assert.NoError(t, err)
		assert.EqualValues(t, len(currentBlock.Messages()), len(currentBlock.MessageResults()))
		assert.EqualValues(t, len(currentBlock.Messages()), len(currentBlock.Receipts()))

		// Verify our method to fetch block hashes works appropriately.
		blockHash, err := chain.BlockHashFromNumber(uint64(i))
		assert.NoError(t, err)
		assert.EqualValues(t, currentBlock.Hash(), blockHash)

		// Try to obtain the state for this block
		_, err = chain.StateAfterBlockNumber(uint64(i))
		assert.NoError(t, err)

		// If we didn't reach genesis, verify our previous block hash matches, and our timestamp is greater.
		if i > 0 {
			previousBlock, err := chain.BlockFromNumber(uint64(i - 1))
			assert.NoError(t, err)
			assert.EqualValues(t, previousBlock.Hash(), currentBlock.Header().ParentHash)
			assert.Less(t, previousBlock.Header().Time, currentBlock.Header().Time)
		}
	}
}

// createChain creates a TestChain used for unit testing purposes and returns the chain along with its initially
// funded accounts at genesis.
func createChain(t *testing.T) (*TestChain, []common.Address) {
	// Create our list of senders
	senders, err := utils.HexStringsToAddresses([]string{
		"0x0707",
		"0x0708",
		"0x0709",
	})
	assert.NoError(t, err)

	// NOTE: Sharing GenesisAlloc between nodes will result in some accounts not being funded for some reason.
	genesisAlloc := make(core.GenesisAlloc)

	// Fund all of our sender addresses in the genesis block
	initBalance := new(big.Int).Div(abi.MaxInt256, big.NewInt(2))
	for _, sender := range senders {
		genesisAlloc[sender] = core.GenesisAccount{
			Balance: initBalance,
		}
	}

	// Create a test chain
	chain, err := NewTestChain(genesisAlloc)
	assert.NoError(t, err)

	return chain, senders
}

// TestChainReverting creates a TestChain and creates blocks and later reverts backward through all possible steps
// to ensure no error occurs and the chain state is restored.
func TestChainReverting(t *testing.T) {
	// Define our probability of jumping
	const blockNumberJumpProbability = 0.20
	const blockNumberJumpMin = 1
	const blockNumberJumpMax = 100
	const blocksToProduce = 50

	// Obtain our chain and senders
	chain, _ := createChain(t)
	chainBackups := make([]*TestChain, 0)

	// Create some empty blocks and ensure we can get our state for this block number.
	for x := 0; x < blocksToProduce; x++ {
		// Determine if this will jump a certain number of blocks.
		if rand.Float32() >= blockNumberJumpProbability {
			// We decided not to jump, so we commit a block with a normal consecutive block number.
			_, err := chain.CreateNewBlock()
			assert.NoError(t, err)
		} else {
			// We decided to jump, so we commit a block with a random jump amount.
			newBlockNumber := chain.HeadBlockNumber()
			jumpDistance := (rand.Uint64() % (blockNumberJumpMax - blockNumberJumpMin)) + blockNumberJumpMin
			newBlockNumber += jumpDistance

			// Determine the jump amount, each block must have a unique timestamp, so we ensure it advanced by at least
			// the diff.

			// Create a block with our parameters
			_, err := chain.CreateNewBlockWithParameters(newBlockNumber, chain.Head().Header().Time+jumpDistance)
			assert.NoError(t, err)
		}

		// Clone our chain
		backup, err := chain.Clone()
		assert.NoError(t, err)
		chainBackups = append(chainBackups, backup)
	}

	// Our chain backups should be in chronological order, so we loop backwards through them and test reverts.
	for i := len(chainBackups) - 1; i >= 0; i-- {
		// Alias our chain backup
		chainBackup := chainBackups[i]

		// Revert our main chain to this block height.
		err := chain.RevertToBlockNumber(chainBackup.HeadBlockNumber())
		assert.NoError(t, err)

		// Verify state matches
		// Verify both chains
		verifyChain(t, chain)
		verifyChain(t, chainBackup)

		// Verify our final block hashes equal in both chains.
		assert.EqualValues(t, chainBackup.Head().Hash(), chain.Head().Hash())
		assert.EqualValues(t, chainBackup.Head().Header().Hash(), chain.Head().Header().Hash())
		assert.EqualValues(t, chainBackup.Head().Header().Root, chain.Head().Header().Root)
	}
}

// TestChainBlockNumberJumping creates a TestChain and creates blocks with block numbers which jumped (are
// non-consecutive) to ensure the chain appropriately spoofs intermediate blocks.
func TestChainBlockNumberJumping(t *testing.T) {
	// Define our probability of jumping
	const blockNumberJumpProbability = 0.20
	const blockNumberJumpMin = 1
	const blockNumberJumpMax = 100
	const blocksToProduce = 200

	// Obtain our chain and senders
	chain, _ := createChain(t)

	// Create some empty blocks and ensure we can get our state for this block number.
	for x := 0; x < blocksToProduce; x++ {
		// Determine if this will jump a certain number of blocks.
		if rand.Float32() >= blockNumberJumpProbability {
			// We decided not to jump, so we commit a block with a normal consecutive block number.
			_, err := chain.CreateNewBlock()
			assert.NoError(t, err)
		} else {
			// We decided to jump, so we commit a block with a random jump amount.
			newBlockNumber := chain.HeadBlockNumber()
			jumpDistance := (rand.Uint64() % (blockNumberJumpMax - blockNumberJumpMin)) + blockNumberJumpMin
			newBlockNumber += jumpDistance

			// Determine the jump amount, each block must have a unique timestamp, so we ensure it advanced by at least
			// the diff.

			// Create a block with our parameters
			_, err := chain.CreateNewBlockWithParameters(newBlockNumber, chain.Head().Header().Time+jumpDistance)
			assert.NoError(t, err)
		}
	}

	// Clone our chain
	recreatedChain, err := chain.Clone()
	assert.NoError(t, err)

	// Verify both chains
	verifyChain(t, chain)
	verifyChain(t, recreatedChain)

	// Verify our final block hashes equal in both chains.
	assert.EqualValues(t, chain.Head().Hash(), recreatedChain.Head().Hash())
	assert.EqualValues(t, chain.Head().Header().Hash(), recreatedChain.Head().Header().Hash())
	assert.EqualValues(t, chain.Head().Header().Root, recreatedChain.Head().Header().Root)
}

// TestChainDynamicDeployments creates a TestChain, deploys a contract which dynamically deploys another contract,
// and ensures that both contract deployments were detected by the TestChain. It also creates empty blocks it
// verifies have no registered contract deployments.
func TestChainDynamicDeployments(t *testing.T) {
	// Copy our testdata over to our testing directory
	contractPath := testutils.CopyToTestDirectory(t, "testdata/contracts/deployment_with_inner.sol")

	// Execute our tests in the given test path
	testutils.ExecuteInDirectory(t, contractPath, func() {
		// Create a solc provider
		solc := platforms.NewSolcCompilationConfig(contractPath)

		// Obtain our compilations and ensure we didn't encounter an error
		compilations, _, err := solc.Compile()
		assert.NoError(t, err)
		assert.EqualValues(t, 1, len(compilations))
		assert.EqualValues(t, 1, len(compilations[0].Sources))

		// Obtain our chain and senders
		chain, senders := createChain(t)

		// Deploy each contract that has no construct arguments.
		deployCount := 0
		for _, compilation := range compilations {
			for _, source := range compilation.Sources {
				for _, contract := range source.Contracts {
					contract := contract
					if len(contract.Abi.Constructor.Inputs) == 0 {
						// Deploy the currently indexed contract
						_, block, err := chain.DeployContract(&contract, senders[0])
						assert.NoError(t, err)
						deployCount++

						// There should've been two address deployments, an outer and inner deployment.
						// (tx deployment and dynamic deployment).
						assert.EqualValues(t, 1, len(block.MessageResults()))
						assert.EqualValues(t, 2, len(block.MessageResults()[0].DeployedContractBytecodes))

						// Ensure we could get our state
						_, err = chain.StateAfterBlockNumber(chain.HeadBlockNumber())
						assert.NoError(t, err)

						// Create some empty blocks and ensure we can get our state for this block number.
						for x := 0; x < 5; x++ {
							block, err = chain.CreateNewBlock()
							assert.NoError(t, err)

							// Empty blocks should not record message results or dynamic deployments.
							assert.EqualValues(t, 0, len(block.MessageResults()))

							_, err = chain.StateAfterBlockNumber(chain.HeadBlockNumber())
							assert.NoError(t, err)
						}
					}
				}
			}
		}

		// Clone our chain
		recreatedChain, err := chain.Clone()
		assert.NoError(t, err)

		// Verify both chains
		verifyChain(t, chain)
		verifyChain(t, recreatedChain)

		// Verify our final block hashes equal in both chains.
		assert.EqualValues(t, chain.Head().Hash(), recreatedChain.Head().Hash())
		assert.EqualValues(t, chain.Head().Header().Hash(), recreatedChain.Head().Header().Hash())
		assert.EqualValues(t, chain.Head().Header().Root, recreatedChain.Head().Header().Root)
	})
}

// TestChainCloning creates a TestChain, sends some messages to it, then clones it into a new instance and ensures
// that the ending state is the same.
func TestChainCloning(t *testing.T) {
	// Copy our testdata over to our testing directory
	contractPath := testutils.CopyToTestDirectory(t, "testdata/contracts/deployment_single.sol")

	// Execute our tests in the given test path
	testutils.ExecuteInDirectory(t, contractPath, func() {
		// Create a solc provider
		solc := platforms.NewSolcCompilationConfig(contractPath)

		// Obtain our compilations and ensure we didn't encounter an error
		compilations, _, err := solc.Compile()
		assert.NoError(t, err)
		assert.True(t, len(compilations) > 0)

		// Obtain our chain and senders
		chain, senders := createChain(t)

		// Deploy each contract that has no construct arguments 10 times.
		for _, compilation := range compilations {
			for _, source := range compilation.Sources {
				for _, contract := range source.Contracts {
					contract := contract
					if len(contract.Abi.Constructor.Inputs) == 0 {
						for i := 0; i < 10; i++ {
							// Deploy the currently indexed contract
							_, _, err = chain.DeployContract(&contract, senders[0])
							assert.NoError(t, err)

							// Ensure we could get our state
							_, err = chain.StateAfterBlockNumber(chain.HeadBlockNumber())
							assert.NoError(t, err)

							// Create some empty blocks and ensure we can get our state for this block number.
							for x := 0; x < i; x++ {
								_, err = chain.CreateNewBlock()
								assert.NoError(t, err)

								_, err = chain.StateAfterBlockNumber(chain.HeadBlockNumber())
								assert.NoError(t, err)
							}
						}
					}
				}
			}
		}

		// Clone our chain
		recreatedChain, err := chain.Clone()
		assert.NoError(t, err)

		// Verify both chains
		verifyChain(t, chain)
		verifyChain(t, recreatedChain)

		// Verify our final block hashes equal in both chains.
		assert.EqualValues(t, chain.Head().Hash(), recreatedChain.Head().Hash())
		assert.EqualValues(t, chain.Head().Header().Hash(), recreatedChain.Head().Header().Hash())
		assert.EqualValues(t, chain.Head().Header().Root, recreatedChain.Head().Header().Root)
	})
}

// TestCallSequenceReplayMatchSimple creates a TestChain, sends some messages to it, then creates another chain which
// it replays the same sequence on. It ensures that the ending state is the same.
// Note: this does not set block timestamps or other data that might be non-deterministic.
// This does not test replaying with a previous call sequence with different timestamps, etc. It expects the TestChain
// semantics to be the same whenever run with the same messages being sent for all the same blocks.
func TestChainCallSequenceReplayMatchSimple(t *testing.T) {
	// Copy our testdata over to our testing directory
	contractPath := testutils.CopyToTestDirectory(t, "testdata/contracts/deployment_single.sol")

	// Execute our tests in the given test path
	testutils.ExecuteInDirectory(t, contractPath, func() {
		// Create a solc provider
		solc := platforms.NewSolcCompilationConfig(contractPath)

		// Obtain our compilations and ensure we didn't encounter an error
		compilations, _, err := solc.Compile()
		assert.NoError(t, err)
		assert.True(t, len(compilations) > 0)

		// Obtain our chain and senders
		chain, senders := createChain(t)

		// Deploy each contract that has no construct arguments 10 times.
		for _, compilation := range compilations {
			for _, source := range compilation.Sources {
				for _, contract := range source.Contracts {
					contract := contract
					if len(contract.Abi.Constructor.Inputs) == 0 {
						for i := 0; i < 10; i++ {
							// Deploy the currently indexed contract
							_, _, err = chain.DeployContract(&contract, senders[0])
							assert.NoError(t, err)

							// Ensure we could get our state
							_, err = chain.StateAfterBlockNumber(chain.HeadBlockNumber())
							assert.NoError(t, err)

							// Create some empty blocks and ensure we can get our state for this block number.
							for x := 0; x < i; x++ {
								_, err = chain.CreateNewBlock()
								assert.NoError(t, err)

								_, err = chain.StateAfterBlockNumber(chain.HeadBlockNumber())
								assert.NoError(t, err)
							}
						}
					}
				}
			}
		}

		// Create another test chain which we will recreate our state from.
		recreatedChain, err := NewTestChainWithGenesis(chain.genesisDefinition)
		assert.NoError(t, err)

		// Replay all messages after genesis
		for i := 1; i < len(chain.blocks); i++ {
			_, err := recreatedChain.CreateNewBlock(chain.blocks[i].Messages()...)
			assert.NoError(t, err)
		}

		// Verify both chains
		verifyChain(t, chain)
		verifyChain(t, recreatedChain)

		// Verify our final block hashes equal in both chains.
		assert.EqualValues(t, chain.Head().Hash(), recreatedChain.Head().Hash())
		assert.EqualValues(t, chain.Head().Header().Hash(), recreatedChain.Head().Header().Hash())
		assert.EqualValues(t, chain.Head().Header().Root, recreatedChain.Head().Header().Root)
	})
}

package blockchain

import (
	"blockchain/pkg/sha"
	"blockchain/pkg/utils"
	"bytes"
	"log"
	"math"
	"math/big"
)

// NOTE take the data from the block
// NOTE create a nonce
// NOTE create a hash of the data + the counter
// NOTE validates hash. Should met requirements

// NOTE Requirements - first few bytes should contain 0s

// NOTE architecture structure to work over same instances
const Diff = 12

type ProfOW struct {
	Block  *Block
	Target *big.Int
}

// NOTE first part of an algorithm, we tale a blocks hash
// NOTE and compare it with `target`, to create a new hash
// NOTE Just like in hash generation function in `blockchain`
func NewProof(b *Block) *ProfOW {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-Diff))

	return &ProfOW{
		b,
		target,
	}
}

func (p *ProfOW) InitData(nonce int) []byte {
	data := bytes.Join([][]byte{p.Block.Hash, p.Block.PrevHash, utils.ToHex(int64(nonce)), utils.ToHex(Diff)}, []byte{})

	return data
}

// RUNS a prof of work
func (p *ProfOW) Run() (int, []byte) {
	var InitNumber big.Int
	var hash [32]byte

	// to indicate difficulty
	nonce := 0

	for nonce < math.MaxInt64 {
		// NOTE infinite loop which runs until hash is met
		// NOTE We prepare data
		// NOTE Hash it into sha256
		// NOTE Convert that hash into BigInt
		// NOTE Compare that int with target
		// NOTE Which is in pow structure,
		// NOTE so we will work over the same block
		data := p.InitData(nonce)
		hash = sha.ComputeHash(data)

		InitNumber.SetBytes(hash[:])

		log.Println(hash)

		// NOTE compare prof of work of target and
		// NOTE new BigInt version of hash
		// NOTE exit condition
		if InitNumber.Cmp(p.Target) == -1 {
			break
		} else {
			nonce++
		}
	}

	return nonce, hash[:]
}

func (pow *ProfOW) Validate() bool {
	var intHash big.Int

	// NOTE compare the nonce with the settled one
	// NOTE return true/false
	data := pow.InitData(pow.Block.Nonce)
	hash := sha.ComputeHash(data)

	intHash.SetBytes(hash[:])

	// compare the hashes
	return intHash.Cmp(pow.Target) == -1
}

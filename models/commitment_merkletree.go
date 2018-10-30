package models

import (
	"errors"
	"fmt"
	"math"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

func printMerkleTree(tree []*chainhash.Hash) {
	num := len(tree)/2 + 1
	i := 0
	for _, hash := range tree {
		if i == num {
			i = 0
			num = num / 2
			fmt.Printf("\n")
		}
		i += 1
		fmt.Printf("%v ", hash)
	}
	fmt.Printf("\n")
}

func buildMerkleTree(hashes []chainhash.Hash) []*chainhash.Hash {
	// Calculate how many entries are required to hold the binary merkle
	// tree as a linear array and create an array of that size.
	nextPoT := nextPow(len(hashes))
	arraySize := nextPoT*2 - 1
	merkles := make([]*chainhash.Hash, arraySize)

	// Create the base transaction hashes and populate the array with them.
	for i := range hashes {
		merkles[i] = &hashes[i]
	}

	// Start the array offset after the last transaction and adjusted to the
	// next power of two.
	offset := nextPoT
	for i := 0; i < arraySize-1; i += 2 {
		switch {
		// When there is no left child node, the parent is nil too.
		case merkles[i] == nil:
			merkles[offset] = nil

		// When there is no right child, the parent is generated by
		// hashing the concatenation of the left child with itself.
		case merkles[i+1] == nil:
			newHash := hashLeaves(*merkles[i], *merkles[i])
			merkles[offset] = newHash

		// The normal case sets the parent node to the double sha256
		// of the concatentation of the left and right children.
		default:
			newHash := hashLeaves(*merkles[i], *merkles[i+1])
			merkles[offset] = newHash
		}
		offset++
	}

	return merkles
}

func hashLeaves(left chainhash.Hash, right chainhash.Hash) *chainhash.Hash {
	// Concatenate the left and right nodes.
	var hash [chainhash.HashSize * 2]byte
	copy(hash[:chainhash.HashSize], left[:])
	copy(hash[chainhash.HashSize:], right[:])

	newHash := chainhash.DoubleHashH(hash[:])
	return &newHash
}

func nextPow(n int) int {
	// Return the number if it's already a power of 2.
	if n&(n-1) == 0 {
		return n
	}

	// Figure out and return the next power of two.
	exponent := uint(math.Log2(float64(n))) + 1
	return 1 << exponent // 2^exponent
}

// CommitmentMerkleTree structure
type CommitmentMerkleTree struct {
	numOfLeaves int
	commitments []chainhash.Hash
	treeStore   []*chainhash.Hash
	root        *chainhash.Hash
}

// New CommitmentMerkleTree instance
// Takes as input a list of commitments and stores these
// along with the whole merkle tree in a list
func NewCommitmentMerkleTree(commitments []chainhash.Hash) CommitmentMerkleTree {
	leavesSize := len(commitments)
	myCommitments := make([]chainhash.Hash, leavesSize)
	copy(myCommitments, commitments)

	treeSize := 2 * nextPow(leavesSize-1)
	myTreeStore := make([]*chainhash.Hash, treeSize)
	myTreeStore = buildMerkleTree(myCommitments)

	return CommitmentMerkleTree{leavesSize, myCommitments, myTreeStore, myTreeStore[treeSize-1]}
}

// Build commitment merkle tree store from commitment hashes
func (m *CommitmentMerkleTree) updateTreeStore() {
	m.treeStore = buildMerkleTree(m.commitments)
	m.root = m.treeStore[len(m.treeStore)-1]
}

// Get merkle tree commitment hash for a specific position in the tree
func (m CommitmentMerkleTree) getMerkleCommitment(position int) (chainhash.Hash, error) {
	if position >= m.numOfLeaves {
		return chainhash.Hash{}, errors.New(fmt.Sprintf("Position %d out of index for merkle tree number of leaves %d", position, m.numOfLeaves))
	}
	return m.commitments[position], nil
}

// Return the whole list of commitments for the merkle tree
func (m CommitmentMerkleTree) getMerkleCommitments() []chainhash.Hash {
	return m.commitments
}

// TODO:
func (m CommitmentMerkleTree) getMerkleProof(position int) (*interface{}, error) {
	if position >= m.numOfLeaves {
		return nil, errors.New(fmt.Sprintf("Position %d out of index for merkle tree number of leaves %d", position, m.numOfLeaves))
	}
	return nil, nil
}

// TODO:
func (m CommitmentMerkleTree) getMerkleProofs() []interface{} {
	return nil
}

// Return the merkle tree store, including all commitments, intermediary tree nodes and root
func (m CommitmentMerkleTree) getMerkleTree() []*chainhash.Hash {
	return m.treeStore
}

// Get tree merkle root
func (m CommitmentMerkleTree) getMerkleRoot() *chainhash.Hash {
	return m.root
}

package models

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/mongodb/mongo-go-driver/bson"
)

// Build merkle proof for a specific position in the merkle tree
func buildMerkleProof(position int, tree []*chainhash.Hash) CommitmentMerkleProof {

	// check proof commitment is valid
	numOfCommitments := len(tree)/2 + 1
	if position >= numOfCommitments || tree[position] == nil {
		return CommitmentMerkleProof{}
	}

	// add base commitment in proof
	var proof CommitmentMerkleProof
	proof.ClientPosition = int32(position)
	proof.Commitment = *tree[position]

	// find all intermediarey commitment ops
	// iterate through each tree height determining
	// the commitment that needs to be added to the proof
	// along with the operation type (append or not)
	var ops []CommitmentMerkleProofOp
	offset := 0
	depth := numOfCommitments
	depthPosition := position
	proofIndex := position
	for depth > 1 {
		var op CommitmentMerkleProofOp
		if proofIndex%2 == 0 { // left side
			op.Append = true
			if tree[proofIndex+1] == nil { // if nil append self
				op.Commitment = *tree[proofIndex]
			} else {
				op.Commitment = *tree[proofIndex+1]
			}
		} else { // right side
			op.Append = false
			op.Commitment = *tree[proofIndex-1]
		}
		ops = append(ops, op)

		// go to next tree height and depth size
		// halve initial position to get corresponding one in new depth
		offset += depth
		depth /= 2
		depthPosition /= 2
		proofIndex = offset + (depthPosition % depth)
	}
	proof.Ops = ops
	proof.MerkleRoot = *tree[len(tree)-1]
	return proof
}

// Prove a commitment using the merkle proof provided
func proveMerkleProof(proof CommitmentMerkleProof) bool {
	hash := proof.Commitment
	for i := range proof.Ops {
		if proof.Ops[i].Append {
			hash = *hashLeaves(hash, proof.Ops[i].Commitment)
		} else {
			hash = *hashLeaves(proof.Ops[i].Commitment, hash)
		}
	}
	return hash == proof.MerkleRoot
}

// CommitmentMerkleProofOps structure
type CommitmentMerkleProofOp struct {
	Append     bool
	Commitment chainhash.Hash
}

// CommitmentMerkleProofOpsBSON structure for mongoDB
type CommitmentMerkleProofOpBSON struct {
	Append     bool   `bson:"append"`
	Commitment string `bson:"commitment"`
}

// Proof OP field names
const (
	PROOF_OP_APPEND_NAME     = "append"
	PROOF_OP_COMMITMENT_NAME = "commitment"
)

// CommitmentMerkleProof structure
type CommitmentMerkleProof struct {
	MerkleRoot     chainhash.Hash
	ClientPosition int32
	Commitment     chainhash.Hash
	Ops            []CommitmentMerkleProofOp
}

// Implement bson.Marshaler MarshalBSON() method for use with db_mongo interface
func (c CommitmentMerkleProof) MarshalBSON() ([]byte, error) {
	proofBson := CommitmentMerkleProofBSON{MerkleRoot: c.MerkleRoot.String(), ClientPosition: c.ClientPosition, Commitment: c.Commitment.String()}

	var opsBson []CommitmentMerkleProofOpBSON
	for _, op := range c.Ops {
		opsBson = append(opsBson, CommitmentMerkleProofOpBSON{op.Append, op.Commitment.String()})
	}
	proofBson.Ops = opsBson
	return bson.Marshal(proofBson)
}

// Implement bson.Unmarshaler UnmarshalJSON() method for use with db_mongo interface
func (c *CommitmentMerkleProof) UnmarshalBSON(b []byte) error {
	var proofBSON CommitmentMerkleProofBSON
	if err := bson.Unmarshal(b, &proofBSON); err != nil {
		return err
	}
	// TODO : not implemented as not required anywhere at the moment
	return nil
}

// Proof field names
const (
	PROOF_MERKLE_ROOT_NAME     = "merkle_root"
	PROOF_CLIENT_POSITION_NAME = "client_position"
	PROOF_COMMITMENT_NAME      = "commitment"
	PROOF_OPS_NAME             = "ops"
)

// CommitmentMerkleProofBSON structure for mongoDB
type CommitmentMerkleProofBSON struct {
	MerkleRoot     string                        `bson:"merkle_root"`
	ClientPosition int32                         `bson:"client_position"`
	Commitment     string                        `bson:"commitment"`
	Ops            []CommitmentMerkleProofOpBSON `bson:"ops"`
}

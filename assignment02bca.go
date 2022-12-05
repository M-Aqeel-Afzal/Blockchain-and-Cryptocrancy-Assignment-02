package assignment02

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"math/big"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
	//a1 "github.com/M-Aqeel-Afzal/assignment01bca"
)

const difficulty_level = 4 //Determines the difficulty to mine block
var transac []string

type Block string
type Hash [20]byte

var m sync.Mutex

type Hashable interface {
	hash() Hash
}

// Creating the left and right node of Merkel Tree
type Node struct {
	left_node  Hashable
	right_node Hashable
}

type block struct { //block node
	Index     int
	Hash      []byte
	Data      string
	Prev_Hash []byte
	root      Node
}
type Blockchain struct { //block chain
	B_chain []*block
}

type EmptyBlock struct {
}

// Appends the Block in the BlockChain
func AddBlock(b_chain *Blockchain, b *block) {
	b_chain.B_chain = append(b_chain.B_chain, b)
}

// Calculates the Hash of the Block
func CalculateHash(data string, prev []byte) []byte {
	// merging data and prev hash
	head := bytes.Join([][]byte{prev, []byte(data)}, []byte{})
	// creating sha256 hash
	hash32 := sha256.Sum256(head)
	// sha256 returns [32]byte
	fmt.Printf("Header hash: %x\n", hash32)
	return hash32[:]
}

// Displays the Data of Block
func DisplayBlock(b *block) {
	fmt.Printf("Current Block Id: %d\nCurrent Block Name: %s\n Current Block Hash: %x\nPrevious Block Hash: %x\n",
		b.Index,
		b.Data,
		b.Hash,
		b.Prev_Hash,
	)
	fmt.Printf("Curent Block Merkel Tree: \n")
	DisplayMerkelTree(b.root) // Calling Display MerkleTree
}

func MineBlock(hash []byte) []byte { //function to mine the block
	target := big.NewInt(1)                                 //for target
	target = target.Lsh(target, uint(256-difficulty_level)) // perform left shift to adjuest the difficulty level
	fmt.Printf("target: %x\n", target)                      //print the target space
	// this is the value that will be incremented and added to the header hash
	var nonce int64
	for nonce = 0; nonce < math.MaxInt64; nonce++ { // run to the size of max int
		testNum := big.NewInt(0)                                              // creating a test number for mining
		testNum.Add(testNum.SetBytes(hash), big.NewInt(nonce))                // adding the nounce to test number
		testHash := sha256.Sum256(testNum.Bytes())                            //creating the hash of test number
		fmt.Printf("\rhash calculated: %x (nonce used: %d)", testHash, nonce) //printing hash and nounce
		if target.Cmp(testNum.SetBytes(testHash[:])) > 0 {                    //comparing the test number hash if lies in the target space
			fmt.Println("\n<========> Congratulations!----Found <========>")
			return testHash[:] //return the hash
		}
	}

	return []byte{}
}

// Creates the New Block
func NewBlock(id int, data []string, prev []byte, blockname string) *block {
	str1 := ""
	for i := 0; i < len(data); i++ {
		str1 += "  " + data[i]
	}
	root := MerkleTree([]Hashable{Block(data[0]), Block(data[1]), Block(data[2]), Block(data[3])})[0].(Node)
	return &block{
		// block Index
		id,
		// first compute a hash with block's header, then mine it
		MineBlock(CalculateHash(str1, prev)),
		// actual data
		blockname,
		// reference to previous block
		prev,
		root,
	}
}

// Creates the Merkle Tree
func MerkleTree(parts []Hashable) []Hashable {
	var nodes []Hashable
	var i int
	for i = 0; i < len(parts); i += 2 {
		if i+1 < len(parts) {
			nodes = append(nodes, Node{left_node: parts[i], right_node: parts[i+1]})
		} else {
			nodes = append(nodes, Node{left_node: parts[i], right_node: EmptyBlock{}})
		}
	}
	if len(nodes) == 1 {
		return nodes
	} else if len(nodes) > 1 {
		return MerkleTree(nodes)
	}
	return nodes
}

func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

// Hash of the BlockChain
func (b Block) hash() Hash {
	return hash([]byte(b)[:])
}

// Hash Function Genisis Block
func (_ EmptyBlock) hash() Hash {
	return [20]byte{}
}

// Hash Function for Nodes
func (n Node) hash() Hash {
	var left, right [sha1.Size]byte
	left = n.left_node.hash()
	right = n.right_node.hash()
	return hash(append(left[:], right[:]...))
}

// Hash Function for the Data
func hash(data []byte) Hash {
	return sha1.Sum(data)
}

func varifyChain(b *block, bc *Blockchain) bool {
	if len(bc.B_chain) > 0 {
		res1 := bytes.Compare(b.Prev_Hash, bc.B_chain[len(bc.B_chain)-1].Hash)
		if res1 != 0 {
			fmt.Printf("<&&== this block make inconsistent chain means do not exist on longest chain (varifychain)==&&>\n")
			fmt.Printf("invalid block\n")

			return false
		}
	}
	for i := 0; i < len(bc.B_chain); i++ {
		res := bytes.Compare(b.Hash, bc.B_chain[i].Hash)
		if res == 0 {
			fmt.Printf("%x <&&==same block(varifychain)==&&> %x \n", b.Hash, bc.B_chain[i].Hash)
			fmt.Printf("invalid block\n")

			return false
		}
		tran1 := get_transaction(b.root)
		transac = nil
		tran2 := get_transaction(bc.B_chain[i].root)
		transac = nil

		for k := 0; k < 4; k++ {
			for j := 0; j < 4; j++ {
				if tran1[j] == tran2[k] {
					fmt.Printf("%s <&&==duplicate transaction(varifychain)==&&> %s \n", tran1[j], tran2[j])
					fmt.Printf("invalid block\n")
					return false
				}
			}
		}
	}

	fmt.Printf("valid block\n")
	return true
}

// get transactions
func get_transaction(node Node) []string {
	getNode(node, 0)
	return transac
}

// Recursive Function to print the Merkle Tree of current Block
func getNode(node Node, level int) {

	if left, check := node.left_node.(Node); check {
		getNode(left, level+1)
	} else if left, check := node.left_node.(Block); check {
		transac = append(transac, string(left))
	}
	if right, check := node.right_node.(Node); check {
		getNode(right, level+1)
	} else if right, check := node.right_node.(Block); check {
		transac = append(transac, string(right))
	}
}

// Display Merkel Tree
func DisplayMerkelTree(node Node) {
	printNode(node, 0)
}

// Recursive Function to print the Merkle Tree of current Block
func printNode(node Node, level int) {
	fmt.Printf("(level: %d)  transaction hash:  %s %s\n", level, strings.Repeat(" ", level), node.hash())

	if left, check := node.left_node.(Node); check {
		printNode(left, level+1)
	} else if left, check := node.left_node.(Block); check {
		fmt.Printf("(level: %d)  transaction hash: %s %s (transactions: %s)\n", level+1, strings.Repeat(" ", level+1), left.hash(), left)

	}
	if right, check := node.right_node.(Node); check {
		printNode(right, level+1)
	} else if right, check := node.right_node.(Block); check {
		fmt.Printf("(level: %d) transaction hash: %s %s (transactions: %s)\n", level+1, strings.Repeat(" ", level+1), right.hash(), right)

	}
}

type pc_info struct { //for storing the information of other pcs
	pc_port string
	pc_ip   string
}
type PC struct { //Node to hold the blockchain
	blockchain  *Blockchain
	is_miner    bool
	is_bootnode bool
	port        string
	name        string
	ip          string
	other_pcs   []*pc_info
}

type Network struct { //block chain
	AllNodes []*PC
}

// updating info of other pcs in bootstrape node
func New_info(port string, ip string) *pc_info {
	return &pc_info{
		port, ip,
	}
}

// Appends the the PC to the p to p network
func Add_pc(net *Network, pc *PC) {
	net.AllNodes = append(net.AllNodes, pc)
}

// update bootstrap other pc_info
func update_pc_info(bootNode *PC, port string, ip string) {
	bootNode.other_pcs = append(bootNode.other_pcs, New_info(port, ip))
}
func get_pc_info(pc *PC, pc1 *PC) {
	for i := 0; i < len(pc.other_pcs); i++ {
		pc1.other_pcs = append(pc1.other_pcs, pc.other_pcs[i])
	}

}

// Creates the New Block
func NewPC(bc *Blockchain, miner_status bool, port string, name string, ip string, boot bool) *PC {
	return &PC{
		// current blockchain
		bc,
		// miner status
		miner_status,
		//bootnode status
		boot,
		// port of pc
		port,
		// name of pc
		name,
		//ip of pc
		ip,
		nil,
	}
}
func transaction_varification(bc *Blockchain, newtran string, pc string) bool {
	for i := 0; i < len(bc.B_chain); i++ {
		alltrans := get_transaction(bc.B_chain[i].root)
		transac = nil
		for j := 0; j < 4; j++ {
			fmt.Printf("%s  <===> %s \n", newtran, alltrans[j])
			if alltrans[j] == newtran {
				fmt.Printf("%s <&&==duplicate transaction(pruning)==&&> %s \n", alltrans[j], newtran)
				return false
			}
		}
	}
	fmt.Printf("transaction ( %s ) varified by %s as new \n", newtran, pc)
	return true

}

// Displays the information of a pc
func DisplayPCs(pc *PC) {
	fmt.Printf("\n name: %s\nport: %s\n miner status: %t\nip address: %s\nbootstrape status: %t\n\n",
		pc.name,
		pc.port,
		pc.is_miner,
		pc.ip,
		pc.is_bootnode,
	)
	fmt.Printf("Neighbours:\n")
	for i := 0; i < len(pc.other_pcs); i++ {
		fmt.Printf("neighbour %d     port:%s     ip address: %s\n", i, pc.other_pcs[i].pc_port, pc.other_pcs[i].pc_ip)
	}

	fmt.Printf("\n\nblockchain:\n")
	for i := 0; i < len(pc.blockchain.B_chain); i++ {

		DisplayBlock(pc.blockchain.B_chain[i]) //display newly created block
	}
}
func Listen(port string, bc *Blockchain, pc int, prev []byte, miner bool) {
	ln, err := net.Listen("tcp", ":"+port)
	var temp = 0
	var block_id = 0
	transactions := []string{}      //transactions
	transactions_temp := []string{} //transactions

	if err != nil {

		log.Fatal(err)

	}
	for {

		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue

		}

		recvdSlice := make([]byte, 100)
		conn.Read(recvdSlice)

		blocks := "" //block

		temp++
		log.Printf("transaction  %s received from  %s at pc%d \n", string(recvdSlice), conn.RemoteAddr(), pc)
		trans := strconv.Itoa(temp+len(bc.B_chain)*4) + " " + string(recvdSlice)
		if transaction_varification(bc, trans, "PC "+strconv.Itoa(pc)) && miner { // transaction varification
			fmt.Printf("transaction  (%s) added to local list by the pc %d \n", trans, pc)
			transactions = append(transactions, trans)
		}

		for j := 0; j < len(transactions); j++ { // transaction pruning
			if transaction_varification(bc, transactions[j], "PC "+strconv.Itoa(pc)) && miner {
				transactions_temp = append(transactions_temp, transactions[j])
			}

		}

		transactions = transactions_temp
		transactions_temp = nil
		//fmt.Println(transactions)
		temp = len(transactions)
		m.Lock()
		if temp%4 == 0 && temp > 0 && miner {

			b := NewBlock(block_id, transactions, prev, "block "+blocks+strconv.Itoa(block_id)) //new block creation
			str1 := ""
			for j := 0; j < len(transactions); j++ {
				str1 += "  " + transactions[j]
			}

			if varifyChain(b, bc) {

				fmt.Printf("===========================new node added to blockchain=====================================")

				DisplayBlock(b) //display newly created block
				AddBlock(bc, b)
				prev = b.Hash // store the hash of block for future use

			}
			block_id++
			transactions = nil
			temp = 0

		}
		time.Sleep(100 * time.Millisecond)
		m.Unlock()
	}
}

func Dial(ip string, port string, tran string) {
	conn, err := net.Dial("tcp", ip+":"+port)
	if err != nil {

		log.Fatal(err)

	}
	conn.Write([]byte(tran))

}

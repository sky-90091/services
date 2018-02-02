package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/skycoin/skycoin/src/api/cli"
	"github.com/skycoin/skycoin/src/api/webrpc"
	"github.com/skycoin/skycoin/src/wallet"
)

var (
	// rpc client
	rpc = rpcCli("127.0.0.1:6430")
)

func walletFromSeed(seed string, count int) (*wallet.ReadableWallet, error) {
	return wallet.CreateAddresses(wallet.CoinTypeSkycoin, seed, count, false)
}

func rpcCli(addr string) *webrpc.Client {
	return &webrpc.Client{
		Addr: addr,
	}
}

func confirmedBalance(unspent *webrpc.OutputsResult) (wallet.Balance, error) {
	return unspent.Outputs.HeadOutputs.Balance()
}

func unconfirmedBalance(unspent *webrpc.OutputsResult) (wallet.Balance, error) {
	return unspent.Outputs.IncomingOutputs.Balance()
}

func currentBlock() uint64 {
	status, err := rpc.GetStatus()
	if err != nil {
		log.Fatal(err)
	}
	return status.BlockNum
}

func printBalance(address string) {
	unspent, err := rpc.GetUnspentOutputs([]string{address})
	if err != nil {
		log.Fatal(err)
	}
	// log.Printf("unspent:%s", spew.Sdump(unspent))

	balance, err := confirmedBalance(unspent)
	if err != nil {
		log.Fatal(err)
	}

	unconfirmed, err := unconfirmedBalance(unspent)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("address:%s confirmed: coins:%d hours:%d unconfirmed: coins:%d hours:%d",
		address, balance.Coins, balance.Hours, unconfirmed.Coins, unconfirmed.Hours)
}

func sendToAddress(wallet, src, dest string) {
	println("beginning")
	printBalance(dest)

	to := cli.SendAmount{Addr: dest, Coins: 100000} // 0.1 sky
	tx, err := cli.CreateRawTxFromWallet(rpc, wallet, src, []cli.SendAmount{to})
	if err != nil {
		log.Fatal(err)
	}
	// log.Printf("tx:%s", spew.Sdump(tx))

	start := time.Now()
	startBlock := currentBlock()

	txid, err := rpc.InjectTransaction(tx)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("injected txid: %s, block: %d", txid, startBlock)

	for {
		res, err := rpc.GetTransactionByID(txid)
		if err != nil {
			log.Fatal(err)
		}
		// log.Printf("res:%s", spew.Sdump(res))
		printBalance(src)
		printBalance(dest)
		if res.Transaction.Status.Confirmed {
			elapsed := time.Since(start)
			log.Printf("confirmed block: %d in: %q", res.Transaction.Status.BlockSeq, elapsed)
			break
		}
		println(".")
		time.Sleep(time.Second * 1)
	}
}

func main() {
	seed := flag.String("seed", "", "seed for source wallet")
	dest := flag.String("dest", "", "destination address")

	flag.Parse()
	if len(*seed) == 0 {
		println("seed must be provided")
		os.Exit(-1)
	}
	if len(*dest) == 0 {
		println("destination address must be provided")
		os.Exit(-1)
	}

	// create wallet with 1 address
	wallet, err := walletFromSeed(*seed, 1)
	if err != nil {
		log.Fatal(err)
	}
	// log.Printf("wallet:%s", spew.Sdump(wallet))

	address := wallet.Entries[0].Address
	// takes in a -seed="wallet seed" parameter-
	// connects to a running skycoin node (locally)
	// generates first skycoin address generated by the wallet seed and prints it
	// checks that there are coins in the first address generated from the wallet seed
	// prints the balance and unspent outputs for the wallet address
	printBalance(address)
	// end of task 1

	walletFile, err := ioutil.TempFile("", "wlt")
	if err != nil {
		log.Fatal(err)
	}
	// these 2 attributes are required
	wallet.Meta["filename"] = walletFile.Name()
	wallet.Meta["type"] = "deterministic"
	err = wallet.Save(walletFile.Name())
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(walletFile.Name()) // clean up
	log.Println("wallet file:", walletFile.Name())

	/*
		task 2:
		create a transaction
		inject the transaction into the network
		record how many blocks the transaction took
		record the physical time (log in seconds) between transactions and how long 10 transactions took
	*/
	sendToAddress(walletFile.Name(), address, *dest)

	/*
		bal, err := visor.ReadableOutputsToUxBalances(unspent.Outputs.HeadOutputs)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("bal:%s", spew.Sdump(bal))

		// key := wallet.Entries[0].Secret // bal[0].Address
		// tx := cli.NewTransaction(spendOutputs, keys, txOuts)

			inUxs, err := unspent.Outputs.SpendableOutputs().ToUxArray()
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("uxs:%s", spew.Sdump(inUxs))

			// create a transaction
			// inject the transaction into the network
			// record how many blocks the transaction took
			// record the physical time (log in seconds) between transactions and how long 10 transactions took

			wallet.Meta["filename"] = "/tmp/xx"
			wallet.Meta["type"] = "deterministic"
			_wallet, err := wallet.ToWallet()
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("walletf:%s", spew.Sdump(_wallet))
			tx, err := cli.CreateRawTx(rpc, &_wallet, []string{address}, address, []cli.SendAmount{to})
	*/

}
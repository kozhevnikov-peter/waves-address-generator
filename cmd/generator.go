package main

import (
	"flag"
	"fmt"
	"regexp"
	"strings"
	"sync"

	wavesplatform "github.com/wavesplatform/go-lib-crypto"
)

var (
	Separator = strings.Repeat("=", 60)
)

type AddressWithSeed struct {
	Seed    wavesplatform.Seed
	Address wavesplatform.Address
}

func main() {
	count := flag.Uint64("n", 1, "count of addresses to generate")
	threads := flag.Int("j", 1, "number of goroutines to generate")
	testnet := flag.Bool("testnet", false, "generate for testnet")
	template := flag.String("t", ".*", "regexp template")
	flag.Parse()

	// Chain ID
	chain := wavesplatform.MainNet
	if *testnet {
		chain = wavesplatform.TestNet
	}

	fmt.Printf("Waves address generator\n\n")
	fmt.Printf("Network: \t%c\n", chain)
	fmt.Printf("Template: \t%d\n", *template)
	fmt.Printf("Threads: \t%d\n", *threads)
	fmt.Printf("Count: \t\t%d\n", *count)
	fmt.Printf("\n")

	done := make(chan struct{}, 1)
	var addressChannel = make(chan AddressWithSeed, *count)
	re, err := regexp.Compile(*template)
	if err != nil || re == nil {
		panic(fmt.Errorf("regexp is not valid: %w", err))
	}

	go printAddress(&addressChannel, *count, &done)

	var waitGroup sync.WaitGroup
	for i := 0; i < *threads; i++ {
		waitGroup.Add(1)
		var wavesCrypto = wavesplatform.NewWavesCrypto()

		go func() {
			defer waitGroup.Done()

			for {
				select {
				case <-done:
					return
				default:
					seed := wavesCrypto.RandomSeed()
					address := wavesCrypto.AddressFromSeed(seed, chain)
					if re.MatchString(string(address)) {
						addressChannel <- AddressWithSeed{Address: address, Seed: seed}
					}
				}
			}
		}()
	}

	waitGroup.Wait()
	close(addressChannel)
}

func printAddress(channel *chan AddressWithSeed, count uint64, done *chan struct{}) {
	for addressWithSeed := range *channel {
		if count == 0 {
			close(*done)
			return
		}

		fmt.Println(Separator)
		fmt.Printf("№: \t\t%d\n", count)
		fmt.Printf("Address: \t%s\n", addressWithSeed.Address)
		fmt.Printf("Seed: \t\t%s\n", addressWithSeed.Seed)
		fmt.Println(Separator)
		count--
	}
}

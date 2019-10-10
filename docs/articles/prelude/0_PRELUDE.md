# Creating A Broker For Submarine Swaps Prelude

This is the beginning of a series of articles about how I created a broker platform for _Submarine Swaps_ written in Golang using the [`btcutil`](https://godoc.org/github.com/btcsuite/btcutil) library and both the [`lnd`](https://godoc.org/github.com/lightningnetwork/lnd) and [`btcd`](https://godoc.org/github.com/btcsuite/btcd) projects. **It is _NOT_ meant to be a platform used in production**. There may be bugs :beetle:, mistakes and things left to optimize. The project's sole purpose was for me to learn more about Golang and find out how to integrate Bitcoin and the Lightning Network in a learning-by-doing approach. The project is live [here](https://noproject.yet) and can only be used with testnet-bitcoins.

By documenting each step of the development I hope that by the end of this article series you will also be able to figure out how to use Golang to interact with the Bitcoin and Lightning network and how to integrate them into your own projects. If you want to get straight into the implementation you can skip over this prelude and go directly to Part 1: Setting Up The Webserver or Part 2: Decoding Lightning Invoices.

Following are some definitions and topics to set the stage for building the Submarine Swaps project.


## Bitcoin

Bitcoin probably does not need a lot of introduction at this point as nearly everyone has heard about it and has some idea about what it may or may not be.

![I don't know how Bitcoin works](images/meme1.jpg)

Just to be sure we are on the same page, here is a quote from [Wikipedia][1]:

> Bitcoin (₿) is a cryptocurrency, a form of electronic cash. It is a decentralized digital currency without a central bank or single administrator that can be sent from user to user on the peer-to-peer bitcoin network without the need for intermediaries.

In other words, as its basis serves a composition of different areas in computer science, such as distributed systems and cryptography. The Bitcoin network keeps track of the transfer of bitcoins using a particular distributed ledger called a [blockchain](https://medium.com/coinmonks/what-the-hell-is-blockchain-and-how-does-it-works-simplified-b9372ecc26ef).

For Submarine Swaps you will need to know about the transaction data structure used in Bitcoin. Transactions in Bitcoin contain a list of instructions describing how the transferred bitcoins can be spend again by e.g. the recipient. This list of instructions has an underlying scripting system, which uses a programming language called _Script_. It allows not only to send bitcoins from a sender **A** to a receiver **B**, but to express complex conditions for spending, typically referred to as _smart contracts_.

For a more thorough and technical introduction to Bitcoin refer to the excellent book [Mastering Bitcoin][2] by Andreas M. Antonopolous. Especially Chapter 7 on the Scripting language will provide more than enough information to follow along. You can also take part in Princeton's fantastic course on Bitcoin on [Coursera](https://www.coursera.org/learn/cryptocurrency).

## Lightning Network

One kind of smart contract is a _payment channel_. A payment channel can be viewed as a lump of bitcoins that can only be spend by two parties in cooperation. Usually, to spend bitcoins one would broadcast a transaction to the whole Bitcoin network updating the current state to reflect the transfer of said bitcoins. This is referred to as an _on-chain_ transaction as it makes direct use of the Bitcoin blockchain. The idea of payment channels is to keep transactions spending from a channel unpublished. Instead of letting the whole network know about the current allocation of bitcoins to each party, only the two parties of the channel will constantly keep track of it. This is called _off-chain_ transacting, as the transactions will typically never hit the blockchain. However, the transactions shared between them always denote the aggregated sum of bitcoins transferred between them inside the channel. If at any time the channel must be closed only the last shared transaction has to be broadcast to the whole Bitcoin network, implicitly closing the channel and returning out of the channel the correct amount of bitcoins to the two parties.

The so-called _Lightning network_ makes use of these payment channels in such a way that not only can two parties send bitcoins between each other off-chain, but they create a whole network of payment channels and can send bitcoins between channels using numerous hops in a trustless manner.

To read more about the Lightning Network checkout Elizabeth Stark's [article](https://coincenter.org/entry/what-is-the-lightning-network) and the resources on [dev.lightning.community](https://dev.lightning.community/resources/). For an in-depth technical explanation see Rusty Russel's [blog posts][3] about Lightning.

## Submarine Swaps

![submarine](images/submarine-swaps.gif)

The Lightning network is an impressive workaround to the limited capacity of Bitcoin. A single on-chain transaction is sufficient to enable payments back and forth between parties near-instantly and without bloating the Bitcoin blockchain. However, a channel can become exhausted if payments tend to be sent more heavily towards one side, locking up all funds of the channel to one party.

The naive approach of re-balancing the channel would be to simply close the exhausted channel and open a completely new one with on-chain bitcoins. However, there exist ways of achieving this without first closing the channel &mdash; one of which is _Submarine Swaps_. Submarine Swaps is a smart contract allowing on-chain payments to a broker in exchange for the equivalent paid off-chain. The advantage of Submarine Swaps is that this happens atomic and is done without trusting the broker, hence without counter-party risk. The technical debt and construction of the swaps will be investigated in part 3: [Making Submarine Swaps](3_LIGHTNING_INVOICES.md). Note also that Submarine Swaps can be performed between _different_ cryptocurrencies.

For more immediate background on Submarine Swaps see the Medium article [How Do Submarine Swaps Work][4] by Rokel Rogstad.

## The Broker Platform

The inspiration to the broker platform came from the [`submarineswaps/swaps-service`][5] repository on Github written in Javascript by Alex Bosworth. I however chose to build the platform for Submarine Swaps with the programming language Golang. The limitation of choosing Go is that we regardless also need Javascript to run on the client's browser to generate a recovery private key never exposed to the broker. The details of it will become clear while we develop the platform.

For this web application I have tried to only use the standard library of Golang. The external `btcutil` library is among other used to generate Bitcoin addresses and transactions, and `btcd` and `lnd` to communicate with the Bitcoin and Lightning network.

Those library and projects have become immense projects developed by [Lightning Labs](https://lightning.engineering/). Their documentations are thorough, however I did not always find it straight forward figuring out which of the numerous functions and different representations of e.g. Bitcoin addresses to use. I will therefore present ways to make use of them to the best of my knowledge, but there may be more appropriate ways of doing it.

I hope you will continue reading the article series and find it both entertaining, interesting and relevant! Also don't hesitate to check out the [repository](https://github.com/bjarnemagnussen/go-submarine-swaps) and add issues or pull requests!


# About The Author

My name is Bjarne Magnussen and I have a great passion for Bitcoin. In my free time I have among other developed on the Bitcoin Python library [`bit`][6] to add multisignature contracts, Segwit and optimization features. Now I am interested in learning Go and therefore decided to publicly develop and document the process of programming the Submarine Swaps broker platform.


[1]:https://en.wikipedia.org/wiki/Bitcoin
[2]:https://github.com/bitcoinbook/bitcoinbook
[3]:https://rusty.ozlabs.org/?p=450
[4]:https://medium.com/suredbits/how-do-submarine-swaps-work-907ed0d91498
[5]: https://github.com/submarineswaps/swaps-service
[6]: https://github.com/ofek/bit

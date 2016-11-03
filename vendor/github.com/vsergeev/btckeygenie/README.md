# btckeygenie v1.0.0

btckeygenie is a standalone Bitcoin keypair/address generator written in Go.
btckeygenie generates an ECDSA secp256k1 keypair, dumps the public key in
compressed and uncompressed Bitcoin address, hexadecimal, and base64 formats,
and dumps the private key in Wallet Import Format (WIF), Wallet Import Format
Compressed (WIFC), hexadecimal, and base64 formats.

btckeygenie includes a lightweight Go package called btckey to easily generate
keypairs, and convert them between compressed and uncompressed varieties of
Bitcoin Address, Wallet Import Format, and raw bytes.

See documentation on btckey here: https://godoc.org/github.com/vsergeev/btckeygenie/btckey

Donations are welcome at `15PKyTs3jJ3Nyf3i6R7D9tfGCY1ZbtqWdv` :-)

## Usage

#### Generating a new keypair

    $ btckeygenie
    Bitcoin Address (Compressed)        1GwX827vFH6cc11sE7jKyhUTsRTZbrNBbD
    Public Key Bytes (Compressed)       02EF0D9A8BA1EB52E14DD33AF3C326B9F5B3C50BFE83D1CD94BDD572DC6492D54E
    Public Key Base64 (Compressed)      Au8Nmouh61LhTdM688MmufWzxQv+g9HNlL3VctxkktVO
    
    Bitcoin Address (Uncompressed)      1EEadeAXyPywyP4AbBijtSVEUDrJ6Uze9b
    Public Key Bytes (Uncompressed)     04EF0D9A8BA1EB52E14DD33AF3C326B9F5B3C50BFE83D1CD94BDD572DC6492D54
                                        EE53FB170859899EDA81F0FF13B6D8A070D3EB872CE96DFAF2D4E06689F154868
    Public Key Base64 (Uncompressed)    BO8Nmouh61LhTdM688MmufWzxQv+g9HNlL3VctxkktVO5T+xcIWYme2oHw/xO22KBw0+uHLOlt+vLU4GaJ8VSGg=
    
    Private Key WIFC (Compressed)       L51L6m126TParjMtoscEiY2Fr9rfXCMW2vyhtLd3wRs9aY27WEKR
    Private Key WIF (Uncompressed)      5Kac96tfK167mM27JR9tbLGEnaGnRy3Nz5XX7CF4PJR3rMHPxgN
    Private Key Bytes                   E8547D576CE8A911BF4DE684BE9E8CBF4F438CE31390D7B9C228FEA18D73786C
    Private Key Base64                  6FR9V2zoqRG/TeaEvp6Mv09DjOMTkNe5wij+oY1zeGw=
    $

#### Importing an existing WIF/WIFC

    $ btckeygenie L51L6m126TParjMtoscEiY2Fr9rfXCMW2vyhtLd3wRs9aY27WEKR
    Bitcoin Address (Compressed)        1GwX827vFH6cc11sE7jKyhUTsRTZbrNBbD
    Public Key Bytes (Compressed)       02EF0D9A8BA1EB52E14DD33AF3C326B9F5B3C50BFE83D1CD94BDD572DC6492D54E
    Public Key Base64 (Compressed)      Au8Nmouh61LhTdM688MmufWzxQv+g9HNlL3VctxkktVO
    
    Bitcoin Address (Uncompressed)      1EEadeAXyPywyP4AbBijtSVEUDrJ6Uze9b
    Public Key Bytes (Uncompressed)     04EF0D9A8BA1EB52E14DD33AF3C326B9F5B3C50BFE83D1CD94BDD572DC6492D54
                                        EE53FB170859899EDA81F0FF13B6D8A070D3EB872CE96DFAF2D4E06689F154868
    Public Key Base64 (Uncompressed)    BO8Nmouh61LhTdM688MmufWzxQv+g9HNlL3VctxkktVO5T+xcIWYme2oHw/xO22KBw0+uHLOlt+vLU4GaJ8VSGg=
    
    Private Key WIFC (Compressed)       L51L6m126TParjMtoscEiY2Fr9rfXCMW2vyhtLd3wRs9aY27WEKR
    Private Key WIF (Uncompressed)      5Kac96tfK167mM27JR9tbLGEnaGnRy3Nz5XX7CF4PJR3rMHPxgN
    Private Key Bytes                   E8547D576CE8A911BF4DE684BE9E8CBF4F438CE31390D7B9C228FEA18D73786C
    Private Key Base64                  6FR9V2zoqRG/TeaEvp6Mv09DjOMTkNe5wij+oY1zeGw=
    $

#### Help/Usage

    $ btckeygenie -h
    Usage: btckeygenie [WIF/WIFC]
    
    btckeygenie v1.0.0 - https://github.com/vsergeev/btckeygenie
    $

## Installation

AUR package: <https://aur.archlinux.org/packages/btckeygenie/>

To fetch, build, and install btckeygenie to `$GOPATH/bin`:

    $ go get github.com/vsergeev/btckeygenie

To build btckeygenie locally:

    $ git clone https://github.com/vsergeev/btckeygenie.git
    $ cd btckeygenie
    $ go build

## Issues

Feel free to report any issues, bug reports, or suggestions at github or by email at vsergeev at gmail.

## License

btckeygenie is MIT licensed. See the included `LICENSE` file for more details.


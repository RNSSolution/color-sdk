sudo kill -9 $(sudo lsof -t -i:26657)
rm -rf ~/.color*
make install
colord init usmanpc --chain-id=abc 
echo 12345678 | colorcli keys add validator
echo 12345678 | colord add-genesis-account $(colorcli keys show validator -a) 10000000000uclr
echo 12345678 | colord gentx --name validator
echo 12345678 | colord collect-gentxs
colord start

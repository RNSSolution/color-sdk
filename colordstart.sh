sudo rm -rf ~/.color*
colord init usmanpc --chain-id=abc
colorcli keys add validator
colord add-genesis-account $(colorcli keys show validator -a) 1000000000000uclr
colord gentx --name validator
colord collect-gentxs
colord start

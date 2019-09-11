colord init usmanpc --chain-id=abc
colorcli keys add validator
colord add-genesis-account $(colorcli keys show validator -a) 10000000000color
colord gentx --name validator
colord collect-gentxs
colord start


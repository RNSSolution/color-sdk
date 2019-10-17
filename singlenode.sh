#!/bin/sh
# This is a comment!
echo Starting single node script

rm -rf ~/.color*
colord init usmanpc --chain-id=test
colorcli keys add validator
colord add-genesis-account $(colorcli keys show validator -a) 1000000000uclr
colord gentx --name validator 
colord collect-gentxs
#colord start

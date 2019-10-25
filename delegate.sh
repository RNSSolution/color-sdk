colorcli keys add usman
colorcli tx send $(colorcli keys show usman -a) 10000uclr --fees=2uclr --from $(colorcli keys show validator -a) --chain-id=abc
#colorcli tx staking delegate colorsvaloper1uysha6j03s2zg6np8jxkywwasmup20xkrumsk5 1000uclr --from usman --chain-id=abc
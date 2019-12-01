colorcli tx gov submit-proposal --title="Test Proposal" --description="My awesome proposal" --type="Text" --deposit="10000000000uclr" --fund="50000000000uclr" --cycle=6 --from node1 --chain-id=colors-test-01 --fees=2uclr --home build/node1/colorcli
sleep 5
colorcli tx gov vote 1 Yes --from node2 --fees=2uclr --chain-id="colors-test-01" --home=build/node2/colorcli
colorcli tx gov vote 1 Yes --from node3 --fees=2uclr --chain-id="colors-test-01" --home=build/node3/colorcli
colorcli tx gov vote 1 Yes --from node4 --fees=2uclr --chain-id="colors-test-01" --home=build/node4/colorcli
colorcli tx gov vote 1 Yes --from node1 --fees=2uclr --chain-id="colors-test-01" --home=build/node1/colorcli
colorcli tx gov vote 1 Yes --from node5 --fees=2uclr --chain-id="colors-test-01" --home=build/node5/colorcli
sleep 5
colorcli tx gov submit-proposal --title="Test Proposal" --description="My awesome proposal" --type="Text" --deposit="10000000000uclr" --fund="50000000000uclr" --cycle=6 --from node1 --chain-id=colors-test-01 --fees=2uclr --home build/node1/colorcli
sleep 5

colorcli tx gov vote 2 Yes --from node2 --fees=2uclr --chain-id="colors-test-01" --home=build/node2/colorcli
colorcli tx gov vote 2 Yes --from node3 --fees=2uclr --chain-id="colors-test-01" --home=build/node3/colorcli
colorcli tx gov vote 2 Yes --from node4 --fees=2uclr --chain-id="colors-test-01" --home=build/node4/colorcli
colorcli tx gov vote 2 Yes --from node5 --fees=2uclr --chain-id="colors-test-01" --home=build/node5/colorcli
colorcli tx gov vote 2 Yes --from node1 --fees=2uclr --chain-id="colors-test-01" --home=build/node1/colorcli

sleep 5
colorcli tx gov submit-proposal --title="Test Proposal" --description="My awesome proposal" --type="Text" --deposit="10000000000uclr" --fund="45000000000uclr" --cycle=6 --from node1 --chain-id=colors-test-01 --fees=2uclr --home build/node1/colorcli
sleep 5

colorcli tx gov vote 2 Yes --from node2 --fees=2uclr --chain-id="colors-test-01" --home=build/node2/colorcli
colorcli tx gov vote 2 Yes --from node3 --fees=2uclr --chain-id="colors-test-01" --home=build/node3/colorcli
colorcli tx gov vote 2 Yes --from node4 --fees=2uclr --chain-id="colors-test-01" --home=build/node4/colorcli
colorcli tx gov vote 2 Yes --from node5 --fees=2uclr --chain-id="colors-test-01" --home=build/node5/colorcli
colorcli tx gov vote 2 Yes --from node1 --fees=2uclr --chain-id="colors-test-01" --home=build/node1/colorcli

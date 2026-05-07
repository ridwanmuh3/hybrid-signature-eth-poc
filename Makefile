run-node:
	anvil --port 8540 --block-gas-limit 50000000

run-client:
	cd client && bun dev

compile-contract:
	cd oracle-server && make compile-contract

forge-deploy:
	cd ethereum && ORACLE_ADDRESS=$(ADDR) forge script script/Deploy.s.sol:Deploy \
		--rpc-url http://localhost:8540 --broadcast --private-key $(KEY)

run-deploy:
	cd oracle-server && go run ./scripts/deploy.go
	
run-oracle:
	# Cek apakah argumen ADDR ada
	@if [ -z "$(ADDR)" ]; then echo "❌ Error: ADDR is missing. Usage: make run-oracle ADDR=0x..."; exit 1; fi
	
	# Ganti Contract Address di .env menggunakan nilai dari $(ADDR)
	sed -i 's/^CONTRACT_ADDRESS=.*/CONTRACT_ADDRESS=$(ADDR)/' ./oracle-server/.env 
	cd oracle-server && make run-dev

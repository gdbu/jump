test:
	@for D in */; do cd $$D && go test && cd ../; done
RESULT = termsnake

build:
	go build -o $(RESULT) main.go

run: build
	./$(RESULT)

clean:
	rm ./$(RESULT)

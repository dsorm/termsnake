# termsnake
![GIF](https://raw.githubusercontent.com/dsorm/termsnake/master/446566.cast.gif)

## How to play
- get the game:
  - play with `docker run -it --rm ghcr.io/dsorm/termsnake`(if you have Docker)
  - download the binary from releases - https://github.com/dsorm/termsnake/releases
  - install with `go install github.com/dsorm/termsnake@latest`(if you have Go installed and properly configured)
- enter your desired width and height of the snake board (must be bigger than 5x5 and must fit in your terminal window)
- use the arrow keys to change snake's direction
- eat food (the red thing) to get bigger
- don't eat yourself (the green thing)

## My assumptions about the task requirements:
* real-time input and output (has to run on it's own without any input, and the user has to be fast)
* has to run in a terminal, no fancy windows
* has to output and refresh the state every round
  * I didn't include the dimensions of the snake board, because it doesn't change and is entered by the user

## My implementation
* I tried to keep as close as possible to the requirements, however where it didn't make sense (would be too complicated), I didn't follow them exactly.
  * I assume that according to the requirements, the snake should be represented in the 2D array/slice. It would however be too complicated to keep track of the curvature of the snake, so the snake isn't inside the array, it is a linked list instead.
* If the snake gets to the wall, it automatically climbs out of the opposite side.
* If you bite yourself (the snake is going straight to the right and you suddenly press left arrow to go straight to left), the game will end. This is not a bug, this is by design.
* The screencast has a bug - `head y` sometimes shows wrong number. This is now fixed.

This task took me about 7 hours to complete.
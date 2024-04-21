# Coding Challenge: Build Your Own NTP Client

This is my solution to the coding challenge [John Crickett's Coding Challenges](https://codingchallenges.fyi/challenges/challenge-ntp).

## Setup

1. Clone the repo
1. Run the tool using of the following approaches:

    ```shell
    # Run the tool automatically using the go command
    $ go run .

    # Build a binary and run it manually
    $ go build -o gontpc
    $ ./gontpc
    ```
1. Done!


## Examples

```shell
$ go run . pool.ntp.org
t1	= 3922693541.211305000
t2	= 3922693541.562651112
t3	= 3922693541.564223255
t4	= 3922693541.488325000
delay	= + 0.06413270183838904
offset	= + 0.04973779048305005
```


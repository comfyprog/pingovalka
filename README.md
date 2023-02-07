# pingovalka
simple dashboard that displays if specified hosts are online

![Screenshot](/assets/screen.png)

You specify list of hosts in a YAML-formatted config file, feed it to `pingovalka` and get a webpage showing if your hosts are online. Simple as that.

## Basic usage
```
$ pingovalka -c config.yml
````
An example config.yml is provided.

## Important
Pingovalka uses github.com/go-ping/ping library that attempts to send an "unprivileged" ping via UDP. On Linux, this must be enabled with the following sysctl command:
```
sudo sysctl -w net.ipv4.ping_group_range="0 2147483647"
````

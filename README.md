pingtastic
==========

An abstract internet latency visualization tool, for monitoring the latency of uplinks and further networks in the internet.

Pingtastic downloads the top 1000 websites from Alexa(Actually starts with the top 1 million csv file) and writes the websites to a mysql database.

After writing the data to a MySQL database, you start the daemon which periodically pings and traceroutes each node, and keeps running track of the path and latency.

Pingtastic then graphs this on a 40x25 Grid.

The vision of pingtastic is that patterns in the grid, will reflect issues happening in the global internet from the perspective of your network.

To use:
Put mysql username, password, database and dbip in the config.ini file.


initialize the database by typing:

`pingtastic getAlexa`

run the server by typing:

`pingtastic server`

Pingtastic would probably run best if monitored by an application like [supervisord](http://supervisord.org/)

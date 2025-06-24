# Fail-Silent-Replicated-Token-Manager-with-Atomic-Semantics-
Developed a fail-silent replicated token manager with atomic semantics using Go, gRPC, and YAML-based configuration for distributed consistency.

This is an extension of  Client-Server Token Model built using GRPC, command Line flags, Mutex, and Bash Script 
This project supports replication among multiple servers.
We have a YAML file with all token details including the access points of all the writers and readers respectively.
 
Here Client file sends all the required requests, which are create, drop, write, and read of tokens to the Server file. 
In Server folder there are two files namely Server.go and TokenCalls.go. In Server.go connection is established and in TokenCalls.go all the functionalities for creating, writing, reading, and deleting the tokens. Also, Hash function, and finding the minimum value function are specified in the same file. 

To run the Server files

bash BashScripts/Server1.sh or ./BashScripts/Server1.sh
bash BashScripts/Server2.sh or ./BashScripts/Server2.sh
bash BashScripts/Server3.sh or ./BashScripts/Server3.sh
bash BashScripts/Server4.sh or ./BashScripts/Server4.sh

To run the Client file 

bash BashScripts/Client.sh

Here when one client performs a write, we take the server details to which the write should be called from YAML file. After write is done we are replicating the details to all other servers.
For operation read, we randomly choose one of the reader server and perform the read operation, then we perform replication operation to all other servers.

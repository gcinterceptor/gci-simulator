from modules import Clients, ServerWithGCI, GCI
import simpy
import configparser

SIM_DURATION_SECONDS = 12
env = simpy.Environment()

config = configparser.ConfigParser()
config.read('configfiles/all-main-conf.ini')

gc_conf = config['gc']
gci_conf = config['gci']
server_conf = config['server']
server = ServerWithGCI(env, server_conf, gc_conf, gci_conf)

requests = list()
client_conf = config['clients']
requests_conf = config['request']
clients = Clients(env, server, requests, client_conf, requests_conf)

env.run(until=SIM_DURATION_SECONDS)

print("Heap level: %.5f%%" % server.heap.level)
print("Remaining requests in queue: %i" % len(server.queue.items))
print("Processed requests: %i" % len(requests))
print("GCI executions: %i" % server.gci.times_performed)
print("GC executions: %i" % server.gc.times_performed)
print("GC execution time sum: %.3f seconds" % server.gci.gc_exec_time_sum)
print("Collects performed: %.i" % server.gc.collects_performed)
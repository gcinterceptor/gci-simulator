from modules import Clients, ServerWithGCI, GCI
import simpy

SIM_DURATION_SECONDS = 12
env = simpy.Environment()

server = ServerWithGCI(env)
requests = list()
clients = Clients(env, server, requests)

env.run(until=SIM_DURATION_SECONDS)

print("Heap level: %.5f%%" % server.heap.level)
print("Remaining requests in queue: %i" % len(server.queue.items))
print("Processed requests: %i" % len(requests))
print("GC executions: %i" % server.gc.times_performed)
print("GCI executions: %i" % server.gci.times_performed)
#print time GCing...
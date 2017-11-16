from modules import Clients, ServerWithGCI, GCI
import simpy

SIM_DURATION = 10
env = simpy.Environment()

server = ServerWithGCI(env)
requests = list()
clients = Clients(env, server, requests)

env.run(until=SIM_DURATION)

print("heap %.10f" % server.heap.level)
print("remaining request in queue: %i" % len(server.queue.items))
print("processed requests %.3f" % len(requests))
print("GC exe %.i" % server.gc.times_performed)
print("GCI exe %.i" % server.gci.times_performed)
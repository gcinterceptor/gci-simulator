from modules import Clients, ServerWithGCI, GCI
import simpy

SIM_DURATION = 10
env = simpy.Environment()

server = ServerWithGCI(env)
requests = list()
clients = Clients(env, server, requests)

env.run(until=SIM_DURATION)

print("remaining request in queue: %i" % len(server.queue.items))
print("processed requests %.3f" % len(requests))
from modules.clients import Clients
from modules.server import ServerWithGCI
from modules.garbage import GCI
import simpy

SIM_DURATION = 10
env = simpy.Environment()

gci = GCI(env)
server = ServerWithGCI(env, gci)
requests = list()
clients = Clients(env, server, requests)

env.run(until=SIM_DURATION)

print("remaining request in queue: %i" % len(server.queue.items))
print("Processed value %.3f" % requests[0]._processed_time)
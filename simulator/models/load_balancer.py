from log import get_logger
import simpy

class LoadBalancer(object):

    def __init__(self, env, server, conf, log_path):
        self.env = env
        self.sleep = float(conf['sleep_time'])

        self.servers = [server]
        self.server_availability = {server.id: 0}
        self.queue = simpy.Store(env)  # the queue of requests
        self.remaining_queue = simpy.Store(env)  # the queue of interrupted requests

        self.logger = get_logger(log_path + "/loadbalancer.log", "LOAD BALANCER")

        self.action = self.env.process(self.run())

    def run(self):
        server = 0
        while True:
            if self.server_availability[self.servers[server].id] <= self.env.now:
                if len(self.remaining_queue.items) > 0:
                    request = yield self.remaining_queue.get()
                    yield self.env.process(self.send_to(server, request))

                elif len(self.queue.items) > 0:
                    request = yield self.queue.get()
                    yield self.env.process(self.send_to(server, request))

            server = (server + 1) % len(self.servers)
            yield self.env.timeout(self.sleep)

    def send_to(self, server, request):
        yield self.env.process(self.servers[server].request_arrived(request))

    def add_server(self, server):
        self.servers.append(server)
        self.server_availability[server.id] = 0

    def request_arrived(self, request):
        request.sent_at(self.env.now)
        yield self.queue.put(request)

    def sucess_request(self, request):
        yield self.env.process(request.client.sucess_request(request))

    def shed_request(self, request, server, unavailable_until):
        self.logger.info(" At %.3f, Request was shedded. The server will be unavailable for: %.3f" % (self.env.now, unavailable_until))
        self.server_availability[server.id] = self.env.now + unavailable_until
        yield self.remaining_queue.put(request)

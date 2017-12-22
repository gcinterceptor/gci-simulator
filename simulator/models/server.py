from log import get_logger
import simpy

class LoadBalancer(object):

    def __init__(self, env, conf, server=None, log_path=None):
        self.env = env
        self.sleep = float(conf['sleep_time'])

        self.servers = []
        self.server_availability = {}
        self.add_server(server)

        self.queue = simpy.Store(env)  # the queue of requests
        self.remaining_queue = simpy.Store(env)  # the queue of interrupted requests

        if log_path:
            self.logger = get_logger(log_path + "/loadbalancer.log", "LOAD BALANCER")
        else:
            self.logger = None

        self.requests_arrived = 0
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
        if server:
            self.servers.append(server)
            self.server_availability[server.id] = 0

    def request_arrived(self, request):
        request.sent_at(self.env.now)
        yield self.queue.put(request)
        self.requests_arrived += 1

    def sucess_request(self, request):
        yield self.env.process(request.client.sucess_request(request))

    def shed_request(self, request, server, unavailable_until):
        if self.logger:
            self.logger.info(" At %.3f, Request was shedded. The server will be unavailable for: %.3f" % (self.env.now, unavailable_until))
        self.server_availability[server.id] = self.env.now + unavailable_until
        yield self.remaining_queue.put(request)

class Server(object):

    def __init__(self, env, id, conf, gc_conf, log_path=None):
        self.env = env
        self.id = id
        self.sleep = float(conf['sleep_time'])

        self.queue = simpy.Store(env) # the queue of requests
        self.interrupted_queue = simpy.Store(env)  # the queue of interrupted requests
        self.heap = simpy.Container(env) # our trash heap

        from .garbage import GC
        self.gc = GC(self.env, self, gc_conf, log_path)

        if log_path:
            self.logger = get_logger(log_path + "/server.log", "SERVER")
        else:
            self.logger = None

        self.is_processing = False
        self.processed_requests = 0
        self.requests_arrived = 0
        self.times_interrupted = 0
        self.action = env.process(self.run())

    def run(self):
        try:
            while True:
                if len(self.interrupted_queue.items) > 0:
                    request = yield self.interrupted_queue.get()  # get a request from store
                    if request.done:
                        self.heap.get(request.memory) # remove trash that shouldn't be added...
                    yield self.env.process(self.process_request(request))

                elif len(self.queue.items) > 0:  # check if there is any request to be processed
                    request = yield self.queue.get()  # get a request from store
                    yield self.env.process(self.process_request(request))

                request = None
                yield self.env.timeout(self.sleep)  # wait for...

        except simpy.Interrupt:
            if self.logger:
                self.logger.info(" At %.3f, Server was interrupted" % (self.env.now))

            self.times_interrupted += 1
            if request:
                yield self.interrupted_queue.put(request)

    def process_request(self, request):
        self.is_processing = True
        yield self.env.process(request.run(self.env, self.heap))
        request.attended_at(self.env.now)
        yield self.env.process(request.load_balancer.sucess_request(request))
        self.processed_requests += 1
        self.is_processing = False


    def request_arrived(self, request):
        request.arrived_at(self.env.now)
        yield self.queue.put(request)   # put the request at the end of the queue
        self.requests_arrived += 1

class ServerWithGCI(Server):

    def __init__(self, env, id, conf, gc_conf, gci_conf, log_path=None):
        super().__init__(env, id, conf, gc_conf, log_path)

        from .garbage import GCI
        self.gci = GCI(self.env, self, gci_conf, log_path)

    def process_request(self, request):
        before = self.env.now
        yield self.env.process(super().process_request(request))
        after = self.env.now
        self.gci.requestFinished(before - after)

    def request_arrived(self, request):
        yield self.env.process(self.gci.before(request))
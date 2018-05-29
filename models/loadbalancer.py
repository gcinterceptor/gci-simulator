from .request import Request


class LoadBalancer(object):

    def __init__(self, create_request_rate, env, server=None):
        self.create_request_rate = create_request_rate
        self.env = env

        self.comunication_delay = 0.008 # 8ms
        self.sleep = 0.00001 #  0.01ms

        self.requests = list()
        self.servers = []
        self.add_server(server)

        self.created_requests = 0
        self.shedded_requests = 0
        self.lost_requests = 0
        self.succeeded_requests = 0

        self.server_index = 0
        self.is_available = True
        self.create_request = env.process(self.create_and_forward_requests())

    def create_and_forward_requests(self):
        time_between_each_sending = 1 / self.create_request_rate
        while True:
            request = Request(self.env, self.created_requests, self.env.now, self)
            self.created_requests += 1
            self.forward(request)
            yield self.env.timeout(time_between_each_sending)

    def forward(self, request):
        request.times_forwarded += 1
        self.env.process(self.servers[self.server_index].request_arrived(request))
        self.server_index = (self.server_index + 1) % len(self.servers)

    def shed_request(self, request):
        self.shedded_requests += 1
        yield self.env.timeout(self.comunication_delay)

        if request.times_forwarded == len(self.servers):
            self.lost_requests += 1
            request.finished_at(self.env.now)
            self.requests.append(request)

        else:
            self.forward(request)

    def request_succeeded(self, request):
        request.finished_at(self.env.now)
        self.succeeded_requests += 1
        self.requests.append(request)
        yield self.env.timeout(self.sleep)

    def add_server(self, server):
        if server:
            self.servers.append(server)

from log import get_logger
from .request import Request


class LoadBalancer(object):

    def __init__(self, env, conf, requests_conf, server=None, log_path=None):
        self.env = env
        self.sleep = float(conf['sleep_time'])

        self.requests = list()
        self.servers = []
        self.server_availability = {}
        self.add_server(server)

        self.created_requests = 0
        self.shedded_requests = 0
        self.lost_requests = 0
        self.succeeded_requests = 0

        if log_path:
            self.logger = get_logger(log_path + "/loadbalancer.log", "LOAD BALANCER")
        else:
            self.logger = None

        self.server_index = 0
        self.is_available = True
        self.create_request = env.process(self.create_and_forward_requests(float(conf['max_requests']), int(conf['create_request_rate']), requests_conf, log_path))

    def create_and_forward_requests(self, max_requests, create_request_rate, requests_conf, log_path):
        time_between_each_sending = 1 / create_request_rate
        while self.created_requests < max_requests:
            request = Request(self.created_requests, self.env, self.env.now, self, requests_conf, log_path)
            self.created_requests += 1
            self.env.process(self.forward(self.server_index, request))
            yield self.env.timeout(time_between_each_sending)

    def forward(self, server_index, request):
        yield self.env.process(self.servers[server_index].request_arrived(request))
        self.server_index = (self.server_index + 1) % len(self.servers)

    def shed_request(self, request, server, unavailable_until):
        self.shedded_requests += 1
        if self.logger:
            self.logger.info(" At %.3f, Request was shedded. The server will be unavailable for: %.3f" % (self.env.now, unavailable_until))

        self.server_availability[server.id] = self.env.now + unavailable_until

        server_index, count, forward = self.server_index, 0, False
        while count < len(self.servers) and not forward:
            if self.env.now >= self.server_availability[self.servers[server_index].id]:
                self.env.process(self.forward(server_index, request))
                forward = True
            else:
                server_index = (server_index + 1) % len(self.servers)
            count += 1

        if not forward:
            self.lost_requests += 1

        yield self.env.timeout(self.sleep)

    def request_succeeded(self, request):
        request.finished_at(self.env.now)
        self.succeeded_requests += 1
        self.requests.append(request)
        yield self.env.timeout(self.sleep)

    def add_server(self, server):
        if server:
            self.servers.append(server)
            self.server_availability[server.id] = 0

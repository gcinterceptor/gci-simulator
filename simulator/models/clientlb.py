from log import get_logger


class Request(object):

    def __init__(self, id, env, created_at, load_balancer, conf, log_path=None):
        self.id = id
        self.env = env
        self.created_at = created_at
        self.load_balancer = load_balancer
        self.service_time = float(conf['service_time'])
        self.memory = float(conf['memory'])

        self.done = False

        self._arrived_time = None
        self._attended_time = None
        self._finished_time = None
        self._latency_time = None

        self._time_in_queue = None
        self._time_in_server = None

        if log_path:
            self.logger = get_logger(log_path + "/request.log", "REQUEST")
        else:
            self.logger = None

    def run(self, heap):
        self._time_in_queue = self.env.now - self._arrived_time
        yield self.env.timeout(self.service_time)
        yield heap.put(self.memory)
        self.done = True

    def arrived_at(self, time):
        self._arrived_time = time

    def attended_at(self, time):
        self._attended_time = time
        self._time_in_server = self._attended_time - self._arrived_time

    def finished_at(self, time):
        self._finished_time = time
        self._latency_time = self._finished_time - self.created_at
        if self.logger:
            self.logger.info(" At %.3f, Request %i was finished. Latency: %.3f" % (self.id, self._finished_time, self._latency_time))


class ClientLB(object):

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
        self.create_request = env.process(self.run(float(conf['max_requests']), int(conf['create_request_rate']), requests_conf, log_path))

    def run(self, max_requests, create_request_rate, requests_conf, log_path):
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

        aux, count, forward = self.server_index, 0, False
        while count < len(self.servers) and not forward:
            if self.env.now >= self.server_availability[self.servers[aux].id]:
                self.env.process(self.forward(aux, request))
                forward = True
            else:
                aux = (aux + 1) % len(self.servers)
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

from simulator.util import getLogger
import simpy

class Request(object):

    def __init__(self, created_at, client, load_balancer, conf):
        self.created_at = created_at
        self.client = client
        self.load_balancer = load_balancer

        self.service_time = float(conf['service_time'])
        self.memory = float(conf['memory'])

        self.done = False
        self._sent_time = None
        self._arrived_time = None
        self._latence_time = None

        self.logger = getLogger("../../data/logs/request.log", "REQUEST")

    def run(self, env, heap):
        yield env.timeout(self.service_time)
        yield heap.put(self.memory)

    def sent_at(self, time):
        self._sent_time = time

    def arrived_at(self, time):
        self._arrived_time = time

    def finished_at(self, time):
        self._latence_time = time - self._sent_time
        self.logger.info(" At %.3f, Request was finished. Latence: %.3f" % (time, self._latence_time))

class Clients(object):

    def __init__(self, env, server, requests, conf, requests_conf):
        self.env = env
        self.server = server
        self.requests = requests
        self.sleep_time = float(conf['sleep_time'])

        self.queue = simpy.Store(env)               # the queue of requests

        self.logger = getLogger("../../data/logs/clients.log", "CLIENTS")

        self.action = env.process(self.send_requests())
        self.create_request = env.process(self.create_requests(int(conf['create_request_rate']),float(conf['max_requests']), requests_conf))

    def send_requests(self):
        while True:
            if len(self.queue.items) > 0:           # check if there is any request to be processed
                request = yield self.queue.get()    # get a request from store
                yield self.env.process(self.server.request_arrived(request))

            yield self.env.timeout(self.sleep_time) # wait for...

    def create_requests(self, create_request_time, max_requests, requests_conf):
        count_requests = 1
        while count_requests <= max_requests:
            request = Request(self.env.now, self, self.server, requests_conf)
            yield self.queue.put(request)
            yield self.env.timeout(1.0 / create_request_time)
            count_requests += 1

    def successfully_sent(self, request):
        request.sent_at(self.env.now)
        yield self.env.timeout(0)

    def sucess_request(self, request):
        request.finished_at(self.env.now)
        self.requests.append(request)
        yield self.env.timeout(0)

    def shed_request(self, request, unavailable_until):
        # don't let the client send new requests until unavailable_until
        # put the request in the front of the queue?
        self.logger.info(" At %.3f, Request was shedded. The server will be unavailable for: %.3f" % (self.env.now, unavailable_until))
        yield self.env.timeout(unavailable_until)
        yield self.queue.put(request)

import simpy

class Request(object):
    def __init__(self, created_at, service_time, memory, client):
        self.created_at = created_at
        self.service_time = service_time
        self.memory = memory
        self.client = client
        self.done = False
        self._done_time = None
        self._sent_time = None
        self._processed_time = None

    def run(self, env, heap):
        yield env.timeout(self.service_time)
        yield heap.put(self.memory)
        self.processed_at(env.now)

    def sent_at(self, time):
        self._sent_time = time

    def processed_at(self, time):
        self._processed_time = time

    def done_at(self, time):
        self._done_time = time

class Clients(object):

    def __init__(self, env, server, requests, sleep_time=0.00001, create_request_time=0.01, max_requests=float("inf"), 
                    request_exec_time=0.0035, memory=0.02):
        self.env = env
        self.server = server
        self.requests = requests
        self.sleep_time = sleep_time
        self.queue = simpy.Store(env)               # the queue of requests
        self.action = env.process(self.send_request())
        self.create_request = env.process(self.create_requests(create_request_time, max_requests, request_exec_time, memory))

    def send_request(self):
        while True:
            if len(self.queue.items) > 0:           # check if there is any request to be processed
                request = yield self.queue.get()    # get a request from store
                yield self.env.process(self.server.request_arrived(request))

            yield self.env.timeout(self.sleep_time) # wait for...

    def create_requests(self, create_request_time, max_requests, request_exec_time, memory):
        count_requests = 1
        while count_requests <= max_requests:
            request = Request(self.env.now, request_exec_time, memory, self)
            yield self.queue.put(request)
            yield self.env.timeout(create_request_time)
            count_requests += 1

    def successfully_sent(self, request):
        request.sent_at(self.env.now)
        yield self.env.timeout(0)

    def sucess_request(self, request):
        request.done_at(self.env.now)
        self.requests.append(request)
        yield self.env.timeout(0)

    def shed_request(self, request, unavailable_until):
        # don't let the client send new requests until unavailable_until
        # put the request in the front of the queue?
        yield self.env.timeout(unavailable_until)
        yield self.queue.put(request)

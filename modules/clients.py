import simpy

class Request(object):
    def __init__(self, created_at, duration, memory, client):
        self.created_at = created_at
        self.duration = duration
        self.memory = memory
        self.client = client
        self.done = False
        self._sent_time = 0
        self._done_time = 0

    def run(self, env, heap):
        yield env.timeout(self.duration)
        yield heap.put(self.memory)

    def sent_at(self, time):
        self._sent_time = time

    def done_at(self, time):
        self._done_time = time

class Clients(object):

    def __init__(self, env, server, requests, process_time=0.00001, create_request_rate=0.01, max_requests=float("inf")):
        self.env = env
        self.server = server
        self.requests = requests
        self.process_time = process_time
        self.queue = simpy.Store(env)  # the queue of requests
        self.action = env.process(self.send_request())
        self.create_request = env.process(self.create_request(create_request_rate, max_requests))

    def send_request(self):
        while True:
            if len(self.queue.items) > 0:  # check if there is any request to be processed
                request = yield self.queue.get()  # get a request from store
                print("At %.3f, CLIENTS a client sent a request" % self.env.now)
                yield self.env.process(self.server.request_arrived(request))

            yield self.env.timeout(self.process_time) # wait for...

    def create_request(self, create_request_rate, max_requests):
        """ Create requests """
        count = 1
        while count <= max_requests:
            duration = 0.035
            memory = 0.02
            request = Request(self.env.now, duration, memory, self)
            print("At %.3f, CLIENTS a client made a request" % self.env.now)
            yield self.queue.put(request)
            yield self.env.timeout(create_request_rate)
            count += 1

    def successfully_sent(self, request):
        print("At %.3f, CLIENTS request successfully sent" % self.env.now)
        request.sent_at(self.env.now)
        yield self.env.timeout(0)

    def sucess_request(self, request):
        print("At %.3f, CLIENTS request successful attended" % self.env.now)
        print("At %.3f, CLIENTS Service time: %.3f" % (self.env.now, self.env.now - request._sent_time))
        request.done_at(self.env.now)
        self.requests.append(request)
        yield self.env.timeout(0)

    def refused_request(self, request, unavailable_until):
        print("At %.3f, CLIENTS refused request.Server is no available until %.3f" % (self.env.now, unavailable_until))
        # don't let the client send new requests until unavailable_until
        # put the request in the front of the queue
        yield self.env.timeout(unavailable_until)
        yield self.queue.put(request)

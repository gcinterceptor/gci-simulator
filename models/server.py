import random


class DataIterator(object):

    def __init__(self, data):
        self.data = data
        self.index = random.randint(0, len(data) - 1)

    def next_value(self):
        self.index = (self.index + 1) % len(self.data)
        return self.data[self.index]


class Server(object):

    def __init__(self, env, id, data):
        self.env = env
        self.id = id
        self.data = DataIterator(data)

        self.requests_arrived = 0
        self.processed_requests = 0
        self.requests = list()

        self.is_shedding = False

    def request_arrived(self, request):
        tmp = self.data.next_value()
        status = int(tmp[0])
        request.status = status
        service_time = tmp[1] / 1000.0  # We use seconds as unit of time.
        request.hops.append(str(tmp[1]))

        if status == 503:
            yield self.env.timeout(service_time)
            request.service_time += service_time
            yield self.env.process(request.load_balancer.request_returned(request))

        elif status == 200:
            self.requests_arrived += 1
            request.arrived_at(self.env.now)
            yield self.env.process(self.process_request(request, service_time))

        else:
            raise Exception("INVALID STATUS DATA AT SERVER")

    def process_request(self, request, service_time):
        yield self.env.process(request.run(service_time))
        yield self.env.process(request.load_balancer.request_returned(request))
        self.processed_requests += 1
        self.requests.append(request)

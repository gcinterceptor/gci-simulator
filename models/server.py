
class DataIterator(object):

    def __init__(self, data):
        self.data = data
        self.index = -1  # This -1 ensures the first next_value() has a self.index with value 0.

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
        status = tmp[0]
        service_time = tmp[1]

        if status == "503":
            request.service_time += service_time
            yield self.env.process(request.load_balancer.shed_request(request))

        elif status == "200":
            self.requests_arrived += 1
            request.arrived_at(self.env.now)
            yield self.env.process(self.process_request(request, service_time))

        else:
            raise Exception("INVALID LATENCY DATA")

    def process_request(self, request, service_time):
        yield self.env.process(request.run(service_time))
        yield self.env.process(request.load_balancer.request_succeeded(request))
        self.processed_requests += 1
        self.requests.append(request)

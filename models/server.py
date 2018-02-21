import random


class Distribution(object):

    def __init__(self, data):
        self.data = data
        self.max = len(self.data)-1
        self.min = 0

    def next_value(self):
        return self.data[random.randint(self.min, self.max)]


class ServerBaseline(object):

    def __init__(self, env, id, service_time_data):
        self.env = env
        self.id = id
        self.service_time_distribution = Distribution(service_time_data)

        self.requests_arrived = 0
        self.processed_requests = 0
        self.requests = list()

    def request_arrived(self, request):
        self.requests_arrived += 1
        request.arrived_at(self.env.now)
        yield self.env.process(self.process_request(request))

    def process_request(self, request):
        yield self.env.process(request.run(self.service_time_distribution.next_value()))
        yield self.env.process(request.load_balancer.request_succeeded(request))
        self.processed_requests += 1
        self.requests.append(request)


class Reprodution(object):

    def __init__(self, data):
        self.data = data
        self.index = -1

    def next_value(self):
        self.index = (self.index + 1) % len(self.data)
        return self.data[self.index]


class ServerControl(ServerBaseline):

    def __init__(self, env, id, service_time_data, processed_request_data, shedded_request_data):
        super().__init__(env, id, service_time_data)

        self.requests_finished_distribution = Reprodution(processed_request_data)
        self.finished = int(self.requests_finished_distribution.next_value())

        self.requests_shedded_distribution = Reprodution(shedded_request_data)
        self.shedded = int(self.requests_shedded_distribution.next_value())

        self.is_shedding = False

    def request_arrived(self, request):
        if self.is_shedding:
            self.shedded -= 1
            if self.shedded == 0:
                self.is_shedding = False
                self.finished = int(self.requests_finished_distribution.next_value())

            yield self.env.process(request.load_balancer.shed_request(request))

        else:
            yield self.env.process(super().request_arrived(request))
            if not self.is_shedding:
                self.finished -= 1
                if self.finished == 0:
                    self.is_shedding = True
                    self.shedded = int(self.requests_shedded_distribution.next_value())


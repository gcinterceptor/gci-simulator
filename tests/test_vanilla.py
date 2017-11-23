from .context import Clients, LoadBalancer, Server, getConfig
import unittest, simpy, os

os.chdir("..")
if not os.path.isdir("logs"):
    os.mkdir("logs")
os.chdir("tests")

class TestVanilla(unittest.TestCase):

    @classmethod
    def setUp(self):
        self.env = simpy.Environment()

        gc_conf = getConfig('../config/gc.ini', 'gc sleep_time-0.00001 threshold-0.9')
        server_conf = getConfig('../config/server.ini', 'server sleep_time-0.00001')
        loadbalancer_conf = getConfig('../config/loadbalancer.ini', 'loadbalancer sleep_time-0.00001')

        self.log_path = '../logs'

        self.server = Server(self.env, 1, server_conf, gc_conf, self.log_path)
        self.load_balancer = LoadBalancer(self.env, self.server, loadbalancer_conf, self.log_path)

        self.requests = list()

    def test_interaction(self):
        requests_conf = getConfig('../config/request.ini', 'request service_time-0.0035 memory-0.02')
        clients_conf = getConfig('../config/clients.ini', 'clients sleep_time-0.00001 create_request_rate-100 max_requests-1')
        num_requests, request_duration, request_memory = int(clients_conf['max_requests']), float(requests_conf['service_time']), float(requests_conf['memory'])

        sim_duration = self.sim_duration_time(num_requests, request_duration, request_memory)
        self.env_run(sim_duration, clients_conf, requests_conf)
        request = self.requests[0]

        expected = 0.000
        self.assert_almost_equal(expected, request.created_at)

        expected = 0.000
        self.assert_almost_equal(expected, request._sent_time)

        expected = 0.0035
        self.assert_almost_equal(expected, request._finished_time)

    def test_one_request_low_heap(self):
        requests_conf = getConfig('../config/request.ini', 'request service_time-0.0035 memory-0.89')
        clients_conf = getConfig('../config/clients.ini', 'clients sleep_time-0.00001 create_request_rate-100 max_requests-1')
        num_requests, request_duration, request_memory = int(clients_conf['max_requests']), float(requests_conf['service_time']), float(requests_conf['memory'])

        sim_duration = self.sim_duration_time(num_requests, request_duration, request_memory)
        self.env_run(sim_duration, clients_conf, requests_conf)

        self.assertEqual(self.server.gc.times_performed, 0)

    def test_one_request_enough_heap(self):
        requests_conf = getConfig('../config/request.ini', 'request service_time-0.0035 memory-0.9')
        clients_conf = getConfig('../config/clients.ini', 'clients sleep_time-0.00001 create_request_rate-100 max_requests-1')
        num_requests, request_duration, request_memory = int(clients_conf['max_requests']), float(requests_conf['service_time']), float(requests_conf['memory'])

        sim_duration = self.sim_duration_time(num_requests, request_duration, request_memory)
        self.env_run(sim_duration, clients_conf, requests_conf)

        self.assertEqual(self.server.gc.times_performed, 1)

    def test_small_queue_low_heap(self):
        requests_conf = getConfig('../config/request.ini', 'request service_time-0.00035 memory-0.089')
        clients_conf = getConfig('../config/clients.ini',
                                 'clients sleep_time-0.00001 create_request_rate-100 max_requests-10')
        num_requests, request_duration, request_memory = int(clients_conf['max_requests']), float(requests_conf['service_time']), float(requests_conf['memory'])

        sim_duration = self.sim_duration_time(num_requests, request_duration, request_memory)
        self.env_run(sim_duration, clients_conf, requests_conf)

        self.assertEqual(self.server.gc.times_performed, 0)

    def test_small_queue_enough_heap(self):
        requests_conf = getConfig('../config/request.ini', 'request service_time-0.00035 memory-0.0900000000000001') # for some reason, without that 1, self.server.heap.level is equal to 0.8999999999999998...
        clients_conf = getConfig('../config/clients.ini',
                                 'clients sleep_time-0.00001 create_request_rate-100 max_requests-10')
        num_requests, request_duration, request_memory = int(clients_conf['max_requests']), float(requests_conf['service_time']), float(requests_conf['memory'])

        sim_duration = self.sim_duration_time(num_requests, request_duration, request_memory)
        self.env_run(sim_duration, clients_conf, requests_conf)

        self.assertEqual(self.server.gc.times_performed, 1)

    def test_high_queue_low_heap(self):
        requests_conf = getConfig('../config/request.ini', 'request service_time-0.000035 memory-0.0089')
        clients_conf = getConfig('../config/clients.ini',
                                 'clients sleep_time-0.00001 create_request_rate-100 max_requests-100')
        num_requests, request_duration, request_memory = int(clients_conf['max_requests']), float(requests_conf['service_time']), float(requests_conf['memory'])

        sim_duration = self.sim_duration_time(num_requests, request_duration, request_memory)
        self.env_run(sim_duration, clients_conf, requests_conf)

        self.assertEqual(self.server.gc.times_performed, 0)

    def test_high_queue_enough_heap(self):
        requests_conf = getConfig('../config/request.ini', 'request service_time-0.000035 memory-0.009')
        clients_conf = getConfig('../config/clients.ini',
                                 'clients sleep_time-0.00001 create_request_rate-100 max_requests-100')
        num_requests, request_duration, request_memory = int(clients_conf['max_requests']), float(requests_conf['service_time']), float(requests_conf['memory'])

        sim_duration = self.sim_duration_time(num_requests, request_duration, request_memory)
        self.env_run(sim_duration, clients_conf, requests_conf)

        self.assertEqual(self.server.gc.times_performed, 1)

    def sim_duration_time(self, num_requests, request_duration, request_memory, sleep_time=0.02, create_request_rate=0.01):
        return num_requests * (request_duration + request_memory + create_request_rate) + sleep_time

    def assert_almost_equal(self, expected, received, delta=0.0001):
        msg = "Expected value: " + str(expected) + ", received value: " + str(received)
        self.assertAlmostEqual(expected, received, msg=msg, delta=delta)

    def env_run(self, sim_duration, clients_conf, requests_conf):
        Clients(self.env, self.load_balancer, self.requests, clients_conf, requests_conf, self.log_path)
        self.env.run(until=sim_duration)

if __name__ == '__main__':
    unittest.main()





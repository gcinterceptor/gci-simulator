from .context import Clients, LoadBalancer, ServerWithGCI, get_config
import unittest, simpy, os

os.chdir("..")
if not os.path.isdir("logs"):
    os.mkdir("logs")
os.chdir("tests")

class TestGCI(unittest.TestCase):

    @classmethod
    def setUp(self):
        self.env = simpy.Environment()

        gc_conf = get_config('../config/gc.ini', 'gc sleep_time-0.00001 threshold-0.9')
        gci_conf = get_config('../config/gci.ini', 'gci sleep_time-0.00001 threshold-0.7 check_heap-2 initial_eget-0.9')
        server_conf = get_config('../config/server.ini', 'server sleep_time-0.00001')
        loadbalancer_conf = get_config('../config/loadbalancer.ini', 'loadbalancer sleep_time-0.00001')

        self.log_path = '../logs'

        self.server = ServerWithGCI(self.env, 1, server_conf, gc_conf, gci_conf, self.log_path)
        self.load_balancer = LoadBalancer(self.env, self.server, loadbalancer_conf, self.log_path)

        self.requests = list()

    def test_interaction(self):
        requests_conf = get_config('../config/request.ini', 'request service_time-0.0035 memory-0.69')
        clients_conf = get_config('../config/clients.ini',
                                 'clients sleep_time-0.00001 create_request_rate-100 max_requests-1')
        num_requests, request_duration, request_memory = int(clients_conf['max_requests']), float(requests_conf['service_time']), float(requests_conf['memory'])

        sim_duration = self.sim_duration_time(num_requests, request_duration, request_memory)
        self.env_run(sim_duration, clients_conf, requests_conf)
        request = self.requests[0]

        expected = 0.000
        self.assert_almost_equal(expected, request.created_at)
        self.assert_almost_equal(expected, request._sent_time)

        expected = 0.0035
        self.assert_almost_equal(expected, request._finished_time)

    def test_one_request_low_heap(self):
        requests_conf = get_config('../config/request.ini', 'request service_time-0.0035 memory-0.69')
        clients_conf = get_config('../config/clients.ini',
                                 'clients sleep_time-0.00001 create_request_rate-100 max_requests-1')
        num_requests, request_duration, request_memory = int(clients_conf['max_requests']), float(requests_conf['service_time']), float(requests_conf['memory'])

        sim_duration = self.sim_duration_time(num_requests, request_duration, request_memory)
        self.env_run(sim_duration, clients_conf, requests_conf)

        self.assert_equal(self.server.gc.times_performed, 0)
        self.assert_equal(self.server.gci.times_performed, 0)

    def test_one_request_enough_heap(self):
        requests_conf = get_config('../config/request.ini', 'request service_time-0.0035 memory-0.7')
        clients_conf = get_config('../config/clients.ini',
                                 'clients sleep_time-0.00001 create_request_rate-100 max_requests-1')
        num_requests, request_duration, request_memory = int(clients_conf['max_requests']), float(
            requests_conf['service_time']), float(requests_conf['memory'])

        sim_duration = self.sim_duration_time(num_requests, request_duration, request_memory)
        self.env_run(sim_duration, clients_conf, requests_conf)

        self.assert_equal(self.server.gc.times_performed, 0)
        self.assert_equal(self.server.gci.times_performed, 0)

    def test_three_request_low_heap(self):
        requests_conf = get_config('../config/request.ini', 'request service_time-0.0035 memory-0.345')
        clients_conf = get_config('../config/clients.ini',
                                 'clients sleep_time-0.00001 create_request_rate-100 max_requests-2')
        num_requests, request_duration, request_memory = int(clients_conf['max_requests']), float(
            requests_conf['service_time']), float(requests_conf['memory'])

        sim_duration = self.sim_duration_time(num_requests, request_duration, request_memory)
        self.env_run(sim_duration, clients_conf, requests_conf)

        self.assert_equal(self.server.gc.times_performed, 0)
        self.assert_equal(self.server.gci.times_performed, 0)

    def test_three_request_enough_heap(self):
        requests_conf = get_config('../config/request.ini', 'request service_time-0.0035 memory-0.35')
        clients_conf = get_config('../config/clients.ini',
                                 'clients sleep_time-0.00001 create_request_rate-100 max_requests-3')
        num_requests, request_duration, request_memory = int(clients_conf['max_requests']), float(
            requests_conf['service_time']), float(requests_conf['memory'])

        sim_duration = self.sim_duration_time(num_requests, request_duration, request_memory)
        self.env_run(sim_duration, clients_conf, requests_conf)

        self.assert_equal(self.server.gc.times_performed, 0)
        self.assert_equal(self.server.gci.times_performed, 1)

    def test_small_queue_low_heap(self):
        requests_conf = get_config('../config/request.ini', 'request service_time-0.00035 memory-0.069')
        clients_conf = get_config('../config/clients.ini',
                                 'clients sleep_time-0.00001 create_request_rate-100 max_requests-11')
        num_requests, request_duration, request_memory = int(clients_conf['max_requests']), float(
            requests_conf['service_time']), float(requests_conf['memory'])

        sim_duration = self.sim_duration_time(num_requests, request_duration, request_memory)
        self.env_run(sim_duration, clients_conf, requests_conf)

        self.assert_equal(self.server.gc.times_performed, 0)
        self.assert_equal(self.server.gci.times_performed, 0)

    def test_small_queue_enough_heap(self):
        requests_conf = get_config('../config/request.ini', 'request service_time-0.00035 memory-0.07')
        clients_conf = get_config('../config/clients.ini',
                                 'clients sleep_time-0.00001 create_request_rate-100 max_requests-11')
        num_requests, request_duration, request_memory = int(clients_conf['max_requests']), float(
            requests_conf['service_time']), float(requests_conf['memory'])

        sim_duration = self.sim_duration_time(num_requests, request_duration, request_memory)
        self.env_run(sim_duration, clients_conf, requests_conf)

        self.assert_equal(self.server.gc.times_performed, 0)
        self.assert_equal(self.server.gci.times_performed, 1)

    def test_high_queue_low_heap(self):
        requests_conf = get_config('../config/request.ini', 'request service_time-0.000035 memory-0.0069')
        clients_conf = get_config('../config/clients.ini',
                                 'clients sleep_time-0.00001 create_request_rate-100 max_requests-101')
        num_requests, request_duration, request_memory = int(clients_conf['max_requests']), float(
            requests_conf['service_time']), float(requests_conf['memory'])

        sim_duration = self.sim_duration_time(num_requests, request_duration, request_memory)
        self.env_run(sim_duration, clients_conf, requests_conf)

        self.assert_equal(self.server.gc.times_performed, 0)
        self.assert_equal(self.server.gci.times_performed, 0)

    def test_high_queue_enough_heap(self):
        requests_conf = get_config('../config/request.ini', 'request service_time-0.000035 memory-0.007')
        clients_conf = get_config('../config/clients.ini',
                                 'clients sleep_time-0.00001 create_request_rate-100 max_requests-101')
        num_requests, request_duration, request_memory = int(clients_conf['max_requests']), float(
            requests_conf['service_time']), float(requests_conf['memory'])

        sim_duration = self.sim_duration_time(num_requests, request_duration, request_memory)
        self.env_run(sim_duration, clients_conf, requests_conf)

        self.assert_equal(self.server.gc.times_performed, 0)
        self.assert_equal(self.server.gci.times_performed, 1)

    def test_multiples_gci_collects(self):
        requests_conf = get_config('../config/request.ini', 'request service_time-0.001 memory-0.1')
        clients_conf = get_config('../config/clients.ini',
                                 'clients sleep_time-0.01 create_request_rate-100 max_requests-71')
        num_requests, request_duration, request_memory = int(clients_conf['max_requests']), float(
            requests_conf['service_time']), float(requests_conf['memory'])

        self.load_balancer.sleep = 0.01

        sim_duration = self.sim_duration_time(num_requests, request_duration, request_memory)
        self.env_run(sim_duration, clients_conf, requests_conf)

        self.assert_equal(self.server.gc.times_performed, 0)
        self.assert_equal(self.server.gci.times_performed, 10)

    def assert_equal(self, expected, received):
        msg = "Expected value: " + str(expected) + ", received value: " + str(received)
        self.assertEqual(expected, received, msg=msg)

    def assert_almost_equal(self, expected, received, delta=0.0001):
        msg = "Expected value: " + str(expected) + ", received value: " + str(received)
        self.assertAlmostEqual(expected, received, msg=msg, delta=delta)

    def sim_duration_time(self, num_requests, request_duration, request_memory, sleep_time=2, create_request_rate=0.01):
        return num_requests * (request_duration + request_memory + create_request_rate) + sleep_time

    def env_run(self, sim_duration, clients_conf, requests_conf):
        clients = Clients(self.env, self.load_balancer, clients_conf, requests_conf, self.log_path)
        self.requests = clients.requests
        self.env.run(until=sim_duration)

if __name__ == '__main__':
    unittest.main()



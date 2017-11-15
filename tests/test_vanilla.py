import sys            # These two first lines, fixes
sys.path.append("..") # the problem of imports from modeles
import unittest, simpy
from modules import Clients, Server

class TestVanilla(unittest.TestCase):

    @classmethod
    def setUp(self):
        self.env = simpy.Environment()
        self.server = Server(self.env)
        self.requests = list()

    def test_interaction(self):
        num_requests, request_duration, request_memory = 1, 0.035, 0.69
        sim_duration = self.sim_duration_time(num_requests, request_duration, request_memory)
        self.env_run(sim_duration, num_requests, request_duration, request_memory)
        request = self.requests[0]
        expected = 0.000
        self.assert_almost_equal(expected, request.created_at)
        self.assert_almost_equal(expected, request._sent_time)
        expected = 0.035
        self.assert_almost_equal(expected, request._processed_time)
        self.assert_almost_equal(expected, request._done_time)

    def test_one_request_low_heap(self):
        num_requests, request_duration, request_memory = 1, 0.0035, 0.89
        sim_duration = self.sim_duration_time(num_requests, request_duration, request_memory)
        self.env_run(sim_duration, num_requests, request_duration, request_memory)
        self.assertEqual(self.server.gc.times_performed, 0)

    def test_one_request_enough_heap(self):
        num_requests, request_duration, request_memory = 1, 0.0035, 0.9
        sim_duration = self.sim_duration_time(num_requests, request_duration, request_memory)
        self.env_run(sim_duration, num_requests, request_duration, request_memory)
        self.assertEqual(self.server.gc.times_performed, 1)

    def test_small_queue_low_heap(self):
        num_requests, request_duration, request_memory = 10, 0.00035, 0.089
        sim_duration = self.sim_duration_time(num_requests, request_duration, request_memory)
        self.env_run(sim_duration, num_requests, request_duration, request_memory)
        self.assertEqual(self.server.gc.times_performed, 0)

    def test_small_queue_enough_heap(self):
        num_requests, request_duration, request_memory = 10, 0.00035, 0.0900000000001 # for some reason, without that 1, self.server.heap.level is equal to 0.8999999999999998...
        sim_duration = self.sim_duration_time(num_requests, request_duration, request_memory)
        self.env_run(sim_duration, num_requests, request_duration, request_memory)
        self.assertEqual(self.server.gc.times_performed, 1)

    def test_high_queue_low_heap(self):
        num_requests, request_duration, request_memory = 100, 0.000035, 0.0089
        sim_duration = self.sim_duration_time(num_requests, request_duration, request_memory)
        self.env_run(sim_duration, num_requests, request_duration, request_memory)
        self.assertEqual(self.server.gc.times_performed, 0)

    def test_high_queue_enough_heap(self):
        num_requests, request_duration, request_memory = 100, 0.000035, 0.009
        sim_duration = self.sim_duration_time(num_requests, request_duration, request_memory)
        self.env_run(sim_duration, num_requests, request_duration, request_memory)
        self.assertEqual(self.server.gc.times_performed, 1)

    def sim_duration_time(self, num_requests, request_duration, request_memory, sleep_time=0.02, create_request_rate=0.01):
        return num_requests * (request_duration + request_memory + create_request_rate) + sleep_time

    def assert_almost_equal(self, expected, received, delta=0.0001):
        msg = "Expected value: " + str(expected) + ", received value: " + str(received)
        self.assertAlmostEqual(expected, received, msg=msg, delta=delta)

    def env_run(self, sim_duration, max_requests, service_time, memory, create_request_rate=0.01):
        Clients(self.env, self.server, self.requests, create_request_rate=create_request_rate, max_requests=max_requests, service_time=service_time, memory=memory)
        self.env.run(until=sim_duration)


if __name__ == '__main__':
    unittest.main()





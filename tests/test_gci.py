import sys            # These two first lines, fixes
sys.path.append("..") # the problem of imports from modeles
import unittest, simpy
from modules import Clients, ServerWithGCI, GCI

class TestGCI(unittest.TestCase):

    @classmethod
    def setUp(self):
        self.env = simpy.Environment()
        self.server = ServerWithGCI(self.env)
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
        num_requests, request_duration, request_memory = 1, 0.0035, 0.69
        sim_duration = self.sim_duration_time(num_requests, request_duration, request_memory)
        self.env_run(sim_duration, num_requests, request_duration, request_memory)
        self.assertEqual(self.server.gc.times_performed, 0)
        self.assertEqual(self.server.gci.times_performed, 0)

    def test_one_request_enough_heap(self):
        num_requests, request_duration, request_memory = 1, 0.0035, 0.8
        sim_duration = self.sim_duration_time(num_requests, request_duration, request_memory)
        self.env_run(sim_duration, num_requests, request_duration, request_memory)
        self.assertEqual(self.server.gc.times_performed, 0)
        self.assertEqual(self.server.gci.times_performed, 0)

    def test_two_request_low_heap(self):
        num_requests, request_duration, request_memory = 2, 0.0035, 0.345
        sim_duration = self.sim_duration_time(num_requests, request_duration, request_memory)
        self.env_run(sim_duration, num_requests, request_duration, request_memory)
        self.assertEqual(self.server.gc.times_performed, 0)
        self.assertEqual(self.server.gci.times_performed, 0)

    def test_three_request_enough_heap(self):
        num_requests, request_duration, request_memory = 3, 0.0035, 0.35
        sim_duration = self.sim_duration_time(num_requests, request_duration, request_memory)
        self.env_run(sim_duration, num_requests, request_duration, request_memory)
        self.assertEqual(self.server.gc.times_performed, 0)
        self.assertEqual(self.server.gci.times_performed, 1)

    def test_small_queue_low_heap(self):
        num_requests, request_duration, request_memory = 10, 0.00035, 0.069
        sim_duration = self.sim_duration_time(num_requests, request_duration, request_memory)
        self.env_run(sim_duration, num_requests, request_duration, request_memory)
        self.assertEqual(self.server.gc.times_performed, 0)
        self.assertEqual(self.server.gci.times_performed, 0)

    def test_small_queue_enough_heap(self):
        num_requests, request_duration, request_memory = 11, 0.00035, 0.07
        sim_duration = self.sim_duration_time(num_requests, request_duration, request_memory)
        self.env_run(sim_duration, num_requests, request_duration, request_memory)
        self.assertEqual(self.server.gc.times_performed, 0)
        self.assertEqual(self.server.gci.times_performed, 1)

    def test_high_queue_low_heap(self):
        num_requests, request_duration, request_memory = 100, 0.000035, 0.0069
        sim_duration = self.sim_duration_time(num_requests, request_duration, request_memory)
        self.env_run(sim_duration, num_requests, request_duration, request_memory)
        self.assertEqual(self.server.gc.times_performed, 0)
        self.assertEqual(self.server.gci.times_performed, 0)

    def test_high_queue_enough_heap(self):
        num_requests, request_duration, request_memory = 101, 0.000035, 0.007
        sim_duration = self.sim_duration_time(num_requests, request_duration, request_memory)
        self.env_run(sim_duration, num_requests, request_duration, request_memory)
        self.assertEqual(self.server.gc.times_performed, 0)
        self.assertEqual(self.server.gci.times_performed, 1)

    def test_multiples_gci_collects(self):
        num_requests, request_duration, request_memory = 70, 0.000035, 0.1
        sim_duration = self.sim_duration_time(num_requests, request_duration, request_memory, create_request_rate=request_duration)
        self.env_run(sim_duration, num_requests, request_duration, request_memory, create_request_rate=request_duration)
        print(self.server.heap.level)
        self.assertEqual(self.server.gc.times_performed, 0)
        self.assertEqual(self.server.gci.times_performed, 10)

    def sim_duration_time(self, num_requests, request_duration, request_memory, sleep_time=2, create_request_rate=0.01):
        return num_requests * (request_duration + request_memory + create_request_rate) + sleep_time

    def assert_almost_equal(self, expected, received, delta=0.0001):
        msg = "Expected value: " + str(expected) + ", received value: " + str(received)
        self.assertAlmostEqual(expected, received, msg=msg, delta=delta)

    def env_run(self, sim_duration, max_requests, service_time, memory, create_request_rate=0.01):
        Clients(self.env, self.server, self.requests, create_request_rate=create_request_rate, max_requests=max_requests, service_time=service_time, memory=memory)
        self.env.run(until=sim_duration)


if __name__ == '__main__':
    unittest.main()



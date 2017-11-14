import sys            # These two first lines, fixes
sys.path.append("..") # the problem of imports from modeles
import unittest, simpy
from modules import Clients, ServerWithGCI, GCI

class TestGCI(unittest.TestCase):
    @classmethod
    def setUp(self):
        self.env = simpy.Environment()
        self.gci = GCI(self.env)
        self.server = ServerWithGCI(self.env, self.gci)
        self.requests = list()

    def test_not_running_GCI(self):
        sim_duration, num_requests, request_duration, request_memoty = 3, 1, 0.035, 0.69
        self.env_run(sim_duration, num_requests, request_duration, request_memoty)
        request = self.requests[0]

        expected = 0.000
        self.assert_almost_equal(expected, request.created_at)
        self.assert_almost_equal(expected, request._sent_time)

        expected = 0.035
        self.assert_almost_equal(expected, request._processed_time)
        self.assert_almost_equal(expected, request._done_time)

        self.assertEqual(self.server.gc.times_performed, 0)
        self.assertEqual(self.server.gci.times_performed, 0)

    def test_running_GCI(self):
        sim_duration, num_requests, request_duration, request_memoty = 3, 1, 0.035, 0.8
        self.env_run(sim_duration, num_requests, request_duration, request_memoty)
        request = self.requests[0]

        expected = 0.000
        self.assert_almost_equal(expected, request.created_at)
        self.assert_almost_equal(expected, request._sent_time)

        expected = 0.035
        self.assert_almost_equal(expected, request._processed_time)
        self.assert_almost_equal(expected, request._done_time)

        self.assertEqual(self.server.gc.times_performed, 0)
        self.assertEqual(self.server.gci.times_performed, 1)

    def assert_almost_equal(self, expected, received, delta=0.0001):
        msg = "Expected value: " + str(expected) + ", received value: " + str(received)
        self.assertAlmostEqual(expected, received, msg=msg, delta=delta)

    def env_run(self, sim_duration, max_requests, duration, memory):
        Clients(self.env, self.server, self.requests, max_requests=max_requests, duration=duration, memory=memory)
        self.env.run(until=sim_duration)


if __name__ == '__main__':
    unittest.main()



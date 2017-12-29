from .context import ClientLB, ServerBaseline, get_config
import unittest
import simpy


class TestBaseline(unittest.TestCase):

    @classmethod
    def setUp(self):
        self.env = simpy.Environment()

    def _build(self, req_conf, lb_conf):
        gc_conf = get_config('config/gcc.ini', 'gcc sleep_time-0.00001 threshold-0.75 collect_duration-0.0019 delay-1')
        req_conf = get_config('config/request.ini', req_conf)
        lb_conf = get_config('config/clientlb.ini', lb_conf)
        server_conf = get_config('config/server.ini', 'server sleep_time-0.00001')

        self.load_balancer = ClientLB(self.env, lb_conf, req_conf)
        self.server = ServerBaseline(self.env, 1, server_conf, gc_conf)
        self.load_balancer.add_server(self.server)

    def failure_msg(self, expected, received):
        return "Expected value: " + str(expected) + ", received value: " + str(received)

    def assert_equal(self, expected, received):
        self.assertEqual(expected, received, msg=self.failure_msg(expected, received))

    def assert_almost_equal(self, expected, received, delta=0.0001):
        self.assertAlmostEqual(expected, received, msg=self.failure_msg(expected, received), delta=delta)

    def test_high_load(self):
        self._build('request service_time-0.006 memory-0.001666667', 'clientlb sleep_time-0.00001 create_request_rate-150 max_requests-inf')

        self.env.run(3)
        expected = 450
        requests, succeeded_requests = self.load_balancer.requests, self.load_balancer.succeeded_requests
        self.assert_equal(expected, len(requests))
        self.assert_equal(expected, succeeded_requests)

    def test_low_load(self):
        self._build('request service_time-0.028 memory-0.007142857', 'clientlb sleep_time-0.00001 create_request_rate-35 max_requests-inf')

        self.env.run(3)
        expected = 105
        requests, succeeded_requests = self.load_balancer.requests, self.load_balancer.succeeded_requests
        self.assert_equal(expected, len(requests))
        self.assert_equal(expected, succeeded_requests)


if __name__ == '__main__':
    unittest.main()

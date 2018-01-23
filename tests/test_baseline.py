from .context import ClientLB, ServerBaseline, get_config
import unittest
import simpy


class TestBaseline(unittest.TestCase):

    @classmethod
    def setUp(self):
        gc_conf = get_config('../config/gcc.ini', 'gcc sleep_time-0.00001 threshold-0.75 collect_duration-0.25 delay-1')
        req_conf = get_config('../config/request.ini', 'request service_time-0.006 memory-0.001666667')
        lb_conf = get_config('../config/clientlb.ini', 'clientlb sleep_time-0.00001 create_request_rate-150 max_requests-inf')
        server_conf = get_config('../config/server.ini', 'server sleep_time-0.00001')

        self.env = simpy.Environment()
        self.load_balancer = ClientLB(self.env, lb_conf, req_conf)
        self.server = ServerBaseline(self.env, 1, server_conf, gc_conf)
        self.load_balancer.add_server(self.server)

    def failure_msg(self, expected, received):
        return "Received value: " + str(received) + ", expected value: " + str(expected)

    def assert_equal(self, expected, received):
        self.assertEqual(expected, received, msg=self.failure_msg(expected, received))

    def assert_almost_equal(self, expected, received, delta=0.0001):
        self.assertAlmostEqual(expected, received, msg=self.failure_msg(expected, received), delta=delta)

    def test_simulation_flow(self):
        self.env.run(3)
        expected = 450 # 150 requests by second after 3 seconds.
        created_requests = self.load_balancer.created_requests
        self.assert_equal(expected, created_requests)
        requests = self.load_balancer.requests
        succeeded_requests = self.load_balancer.succeeded_requests
        expected = 449 # Should have 450 succeeded requests, but the last one was delayed by a gc execution.
        self.assert_equal(expected, len(requests))
        self.assert_equal(expected, succeeded_requests)

        self.env.run(4)
        expected = 600
        created_requests = self.load_balancer.created_requests
        self.assert_equal(expected, created_requests)

        self.env.run(4.0061)
        expected = 601
        created_requests = self.load_balancer.created_requests
        self.assert_equal(expected, created_requests)
        expected = 451  # since at 3 gc starts gcing, the request duration will be 0.006 time service + 1 delay
        requests = self.load_balancer.requests
        succeeded_requests = self.load_balancer.succeeded_requests
        self.assert_equal(expected, len(requests))
        self.assert_equal(expected, succeeded_requests)

        self.env.run(5.0061)
        expected = 751
        created_requests = self.load_balancer.created_requests
        self.assert_equal(expected, created_requests)
        expected = 616  # 451 (before) + 166 (since 1/0.006 is 166.6...) - 1 (the request delayed by a gc execution.)
        requests = self.load_balancer.requests
        succeeded_requests = self.load_balancer.succeeded_requests
        self.assert_equal(expected, len(requests))
        self.assert_equal(expected, succeeded_requests)

        self.env.run(6.0061)
        expected = 901
        created_requests = self.load_balancer.created_requests
        self.assert_equal(expected, created_requests)
        expected = 782  # 617 (before) + 166 (since 1/0.006 is 166.6...) - 1 (the request delayed by a gc execution.)
        requests = self.load_balancer.requests
        succeeded_requests = self.load_balancer.succeeded_requests
        self.assert_equal(expected, len(requests))
        self.assert_equal(expected, succeeded_requests)

        self.env.run(7.0061)
        expected = 1051
        created_requests = self.load_balancer.created_requests
        self.assert_equal(expected, created_requests)
        expected = 899 # the request number 900 received 1 sec of delay after the request number 899.
        requests = self.load_balancer.requests
        succeeded_requests = self.load_balancer.succeeded_requests
        self.assert_equal(expected, len(requests))
        self.assert_equal(expected, succeeded_requests)

    def test_flags_variables(self):
        self.env.run(7)

        expected = 899 # 900 - 1 (request delayed)
        processed_requests = self.server.processed_requests
        self.assert_equal(expected, processed_requests)
        expected = 1050
        requests_arrived = self.server.requests_arrived
        self.assert_equal(expected, requests_arrived)
        expected = 0
        times_interrupted = self.server.times_interrupted
        self.assert_equal(expected, times_interrupted)

        expected = 2
        times_performed = self.server.gc.times_performed
        self.assert_equal(expected, times_performed)
        collects_performed = self.server.gc.collects_performed
        self.assert_equal(expected, collects_performed)
        expected = 0.5
        gc_exec_time_sum = self.server.gc.gc_exec_time_sum
        self.assert_almost_equal(expected, gc_exec_time_sum)

    def test_request_list(self):
        self.env.run(20)
        requests = self.load_balancer.requests

        id_set = set()
        created_time_set = set()
        arrived_time_set = set()
        attended_time_set = set()
        finished_time_set = set()

        for request in requests:
            self.assertTrue(request.done)
            id_set.add(request.id)
            created_time_set.add(request.created_time)
            arrived_time_set.add(request._arrived_time)
            attended_time_set.add(request._attended_time)
            finished_time_set.add(request._finished_time)

        expected = len(requests)
        self.assert_equal(expected, len(id_set))
        self.assert_equal(expected, len(created_time_set))
        self.assert_equal(expected, len(arrived_time_set))
        self.assert_equal(expected, len(attended_time_set))
        self.assert_equal(expected, len(finished_time_set))


if __name__ == '__main__':
    unittest.main()

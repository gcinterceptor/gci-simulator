from .context import LoadBalancer, ServerBaseline, ServerControl
import unittest
import simpy


class TestLoadBalancer(unittest.TestCase):

    @classmethod
    def setUp(self):
        self.env = simpy.Environment()
        id = 0
        service_time_data = [6]
        self.server = ServerBaseline(self.env, id, service_time_data)
        create_request_rate = 80
        self.lb = LoadBalancer(create_request_rate, self.env, self.server)

    def failure_msg(self, expected, received):
        return "Received value: " + str(received) + ", expected value: " + str(expected)

    def assert_equal(self, expected, received):
        self.assertEqual(expected, received, msg=self.failure_msg(expected, received))

    def test_request(self):
        self.assert_equal(0, self.env.now)
        sim_duration = 6.012500000000001  # time enought to process only two request
        self.env.run(sim_duration)
        self.assert_equal(sim_duration, self.env.now)

        self.assert_equal(2, len(self.lb.requests))

        request = self.lb.requests[0]
        self.assert_equal(6, request.service_time)
        self.assert_equal(6, request._latency)
        self.assert_equal(0, request._arrived_time)

        request = self.lb.requests[1]
        self.assert_equal(6, request.service_time)
        self.assert_equal(6, request._latency)
        self.assert_equal(0.0125, request._arrived_time)

    def test_request_creation(self):
        self.assert_equal(0, self.lb.created_requests)
        self.env.run(50)
        self.assert_equal(4000, self.lb.created_requests)
        self.env.run(100)
        self.assert_equal(8000, self.lb.created_requests)

    def test_forward(self):

        servers = [self.server]
        service_time_data = [6]
        for i in range(1, 4):
            servers.append(ServerBaseline(self.env, i, service_time_data))
            self.lb.add_server(servers[i])

        self.env.run(0.05) # run enough time to send only 4 requests.

        for i in range(4):
            self.assert_equal(1, servers[i].requests_arrived)

        self.env.run(0.2) # run enough time to send only 12 requests.

        for i in range(4):
            self.assert_equal(4, servers[i].requests_arrived)

    def test_shed(self):
        id = 0
        service_time_data = [0.01]
        processed_request_data = [1]
        shedded_request_data = [2]
        self.lb.servers[0] = ServerControl(self.env, id, service_time_data, processed_request_data, shedded_request_data)

        id = 1
        service_time_data = [0.01]
        processed_request_data = [2]
        shedded_request_data = [1]
        self.lb.add_server(ServerControl(self.env, id, service_time_data, processed_request_data, shedded_request_data))

        self.env.run(0.05)

        self.assertTrue(self.lb.requests[0].done)
        self.assertTrue(self.lb.requests[1].done)
        self.assertTrue(self.lb.requests[2].done)
        self.assertFalse(self.lb.requests[3].done)

        self.assert_equal(1, self.lb.requests[0].times_forwarded)
        self.assert_equal(1, self.lb.requests[1].times_forwarded)
        self.assert_equal(2, self.lb.requests[2].times_forwarded)
        self.assert_equal(2, self.lb.requests[3].times_forwarded)

        self.assert_equal(1, self.lb.lost_requests) # only one requests have be shedded by both servers
        self.assert_equal(3, self.lb.shedded_requests) # One request were shedded by one server and processed by the other and one request were shedded by both servers.


if __name__ == '__main__':
    unittest.main()

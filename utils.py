import configparser, logging

def get_config(path_file, conf_section):
    config_parser = configparser.ConfigParser()
    config_parser.read(path_file)
    required_conf = config_parser[conf_section]

    return required_conf

def get_logger(path_file, logger_name):
    handler = logging.FileHandler(path_file, mode='w')

    formatter = logging.Formatter('%(asctime)s %(levelname)s %(message)s')
    handler.setFormatter(formatter)

    logger = logging.getLogger(logger_name)
    logger.setLevel(logging.INFO)
    logger.addHandler(handler)

    return logger

def generate_results(path_file, logger_name, request_list):
    logger = get_logger(path_file, logger_name)

    logger.info("LATENCY:")
    for request in request_list:
        logger.info(request._latency_time)

    logger.info("Number of requests: %.i" % len(request_list))

def flag(env, time):
    yield env.timeout(time)
    print(env.now)
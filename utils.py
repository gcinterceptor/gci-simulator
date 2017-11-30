import configparser, logging, csv

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
    logger.setLevel(logging.DEBUG)
    logger.addHandler(handler)

    return logger

def generate_results(data, path_file):
    with open(path_file, "w", newline='') as csv_file:
        writer = csv.writer(csv_file, delimiter=',')
        for line in data:
            writer.writerow(line)

import configparser

def get_config(path_file, conf_section):
    config_parser = configparser.ConfigParser()
    config_parser.read(path_file)
    required_conf = config_parser[conf_section]

    return required_conf
from .exponential import Exponential 
from .normal import Normal

class Distribution(object):
    
    def __init__(self, distribution_name, parameters):
        self.distribution = None
        
        if(distribution_name == 'exponential'):
            avg = parameters[0]
            self.distribution = Exponential(avg)
            
        elif(distribution_name == 'normal'):
            avg = parameters[0]
            desviation = parameters[1]
            self.distribution = Normal(avg, desviation)
    
    def get_next_value(self):
        return self.distribution.get_next_value()

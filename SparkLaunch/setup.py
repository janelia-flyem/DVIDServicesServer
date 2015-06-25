#from distutils.core import setup
from setuptools import setup, find_packages

setup(name = "SparkLaunch",
    version = "1.0",
    description = "Launch spark cluster job.",
    author = "Stephen Plaza",
    author_email = 'plazas@janelia.hhmi.org',
    license = 'LICENSE.txt',
    packages = [],
    package_data = {},
    install_requires = [ ],
    scripts = ["bin/spark_launch", "bin/spark_launch_wrapper"]
)

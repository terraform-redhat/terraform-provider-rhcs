# -*- coding: utf-8 -*-

#
# Copyright (c***REMOVED*** 2019 Red Hat, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License"***REMOVED***;
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

"""
Functions and classes to simplify logging.
"""

import sys


class Logger:
    """
    Simple logger that knows how to redact certain parts of the messages that
    it produces.
    """
    def __init__(self, file=None***REMOVED***:
        """
        Creates a new logger that will write to the given file or to the
        standard output if no file is provided.
        """
        self._redacted = set(***REMOVED***
        if file:
            self._file = open(file, "w"***REMOVED***
        else:
            self._file = sys.stdout

    def _redact(self, message***REMOVED***:
        result = message
        for value in self._redacted:
            result = result.replace(value, "***"***REMOVED***
        return result

    def info(self, message***REMOVED***:
        """
        Writes the given message to the log.
        """
        redacted = self._redact(message***REMOVED***
        print(redacted, file=self._file, flush=True***REMOVED***

    def redact(self, value***REMOVED***:
        """
        Adds the given value to the set of texts that will be redacted from the
        output, replacing each occurence with three asterisks.
        """
        if value != "":
            self._redacted.add(value***REMOVED***

    def close(self***REMOVED***:
        """
        Closes the logger and releases all the resources it was using.
        """
        if self._file != sys.stdout:
            self._file.close(***REMOVED***

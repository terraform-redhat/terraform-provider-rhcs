# -*- coding: utf-8 -*-

#
# Copyright (c) 2019 Red Hat, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
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
    def __init__(self, file=None):
        """
        Creates a new logger that will write to the given file or to the
        standard output if no file is provided.
        """
        self._redacted = set()
        if file:
            self._file = open(file, "w")
        else:
            self._file = sys.stdout

    def _redact(self, message):
        result = message
        for value in self._redacted:
            result = result.replace(value, "***")
        return result

    def info(self, message):
        """
        Writes the given message to the log.
        """
        redacted = self._redact(message)
        print(redacted, file=self._file, flush=True)

    def redact(self, value):
        """
        Adds the given value to the set of texts that will be redacted from the
        output, replacing each occurence with three asterisks.
        """
        if value != "":
            self._redacted.add(value)

    def close(self):
        """
        Closes the logger and releases all the resources it was using.
        """
        if self._file != sys.stdout:
            self._file.close()

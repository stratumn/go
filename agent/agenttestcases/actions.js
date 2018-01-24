// Copyright 2017 Stratumn SAS. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

module.exports = {
    events: {
      didSave: function(segment) {
        console.log("Segment " + segment.meta.linkHash + " was saved!");
      }
    },
  
    name: "test",
  
    init: function(title) {
      if (!title) {
        return this.reject("a title is required");
      }
  
      this.state = {
        title: title
      };
  
      this.meta.tags = [title];
      console.log("now is", this);
      this.append();
    },
  
    test: function(title) {
      if (!title) {
        return this.reject("a title is required");
      }
  
      this.state = {
        title: title
      };
  
      this.append();
    }
  };
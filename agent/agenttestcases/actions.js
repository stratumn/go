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
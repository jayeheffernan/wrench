



 





function console(statement) {
  print(statement + "\r\n")
}

function dbg(fileline, statement) {
  console(fileline + " " + statement)
}











function NewMockMethod(name, n_args, checkcall) {
  if (n_args == 1) {
    return function(arg1) {
      checkcall(name, [arg1])
    }
  } else if (n_args == 2) {
    return function(arg1, arg2) {}
  } else if (n_args == 3) {
    return function(arg1, arg2, arg3) {}
  } else if (n_args == 4) {
    return function(arg1, arg2, arg3, arg4) {}
  }
  return function() {}
}

function NewMockFunction(name, n_args) {
  if (n_args == 1) {
    return function(original_this, arg1) {}
  } else if (n_args == 2) {
    return function(original_this, arg1, arg2) {}
  } else if (n_args == 3) {
    return function(original_this, arg1, arg2, arg3) {}
  } else if (n_args == 4) {
    return function(original_this, arg1, arg2, arg3, arg4) {}
  }
  return function(original_this) {
  }
}

function NewMockObject() {
  return {
    name = "a",
    calls = {},
    
    function CheckMethodCall(method_name, args) {
      local call_list = null;
      console("Function: " + method_name + " called with args: " + args)
      try {
        call_list = this.calls[method_name]
      } catch (e) {
        dbg("sq_unit.nut:45", "Didn't find a list of expected calls...")
      }
      if (call_list == null) {
        dbg("sq_unit.nut:48", "No call list and got call...")
        assert(false)
      }
      
      console("Call list args " + call_list.len() + " args " + args.len())
      if (call_list.len() != args.len()) {
        dbg("sq_unit.nut:54", "Length of expected args and received args is different for function: " + method_name)
        assert(false)
      }
    }
    
    function NewMethod(name, n_args) {
      local f = NewMockMethod(name, n_args, this.CheckMethodCall)
      this[name] <- f.bindenv(this)
      return name
    }
    
    function ExpectMethodCall(name, args) {
      local call_list = null
      try {
        call_list = this.calls[name]
      } catch (e) {
        
      }
      if (call_list == null) {
        this.calls[name] <- []
        call_list = this.calls[name]
      }
      call_list.append(args)
      console("Added call to: " + name + " with args " + args)
    }
  }
}



dbg("tests\blah.test.nut:2", "Blah test start")

console("tests\blah.test.nut:17")
local y = NewMockObject()
local f = y.NewMethod("dude2", 1)
y.ExpectMethodCall(f, [1])
y.ExpectMethodCall(f, [2])
y.dude2(1)
dbg("tests\blah.test.nut:10", "Blah test end")
package org.apache.custos.rest.controller;

import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestMethod;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/cilogon")
public class CILogonController {

    @RequestMapping(value = "hello", method = RequestMethod.GET)
    public String helloMethod() {
        return "Hello";
    }
}

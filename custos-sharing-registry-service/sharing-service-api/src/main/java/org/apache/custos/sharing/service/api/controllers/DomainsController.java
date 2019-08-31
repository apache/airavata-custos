package org.apache.custos.sharing.service.api.controllers;

import org.apache.custos.sharing.service.core.models.Domain;
import org.apache.custos.sharing.service.core.exceptions.SharingRegistryException;
import org.apache.custos.sharing.service.core.service.SharingRegistryService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.Collections;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/domain")
public class DomainsController {

    @Autowired
    private static SharingRegistryService sharingRegistryService;

    /**
     <p>API method to create a new domain</p>
     */
    @RequestMapping(value = "/create", method= RequestMethod.POST)
    public ResponseEntity<String> createDomain(@RequestBody Domain domain){
        String domainId = sharingRegistryService.createDomain(domain);
        if(domainId.equals(domain.getDomainId())) return new ResponseEntity<>(HttpStatus.OK);
        else throw new SharingRegistryException("Domain creation failed");
    }

    /**
     <p>API method to update a domain</p>
     */
    @RequestMapping(value = "/update", method = RequestMethod.PUT)
    public ResponseEntity<String> updateDomain(@RequestBody Domain domain){
        boolean response = sharingRegistryService.updateDomain(domain);
        if(response) return new ResponseEntity<>(HttpStatus.OK);
        else throw new SharingRegistryException("Domain update failed");
    }

    /**
     <p>API method to check Domain Exists</p>
     */
    @RequestMapping(value = "/exists/{domainId}", method = RequestMethod.GET)
    public @ResponseBody
    Map<String, Boolean> isDomainExists(@PathVariable("domainId") String domainId) {
        boolean response = sharingRegistryService.isDomainExists(domainId);
        return Collections.singletonMap("exists", response);
    }

    /**
     <p>API method to retrieve a domain</p>
     */
    @RequestMapping(value = "/{domainId}",method = RequestMethod.GET)
    public ResponseEntity<Domain> getDomain(@PathVariable("domainId") String domainId){
        Domain domain = sharingRegistryService.getDomain(domainId);
        if(domain != null) return new ResponseEntity<>(domain, HttpStatus.OK);
        else return new ResponseEntity<>(HttpStatus.NOT_FOUND);
    }

    /**
     <p>API method to delete domain</p>
     */
    @RequestMapping(value = "/{domainId}", method = RequestMethod.DELETE)
    public @ResponseBody Map<String, Boolean> deleteDomain(@PathVariable("domainId") String domainId){
        boolean response = sharingRegistryService.deleteDomain(domainId);
        return Collections.singletonMap("success", response);
    }

    /**
     <p>API method to get all domain.</p>
     */
    @RequestMapping(value = "/offset/{offset}/limit/{limit}", method = RequestMethod.GET)
    public @ResponseBody List<Domain> getDomains(@PathVariable("offset") int offset, @PathVariable("limit") int limit){
        return sharingRegistryService.getDomains(offset, limit);
    }

}

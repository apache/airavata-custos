package org.apache.custos.sharing.service.api.controllers;

import org.apache.custos.sharing.service.core.models.Domain;
import org.apache.custos.sharing.service.core.service.SharingRegistryService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import javax.validation.Valid;
import java.util.List;

@RestController
@RequestMapping("/domains")
public class DomainsController {

    @Autowired
    SharingRegistryService sharingRegistryService;

    /**
     <p>API method to create a new domain</p>
     */
    @RequestMapping(method= RequestMethod.POST)
    public ResponseEntity<Domain> createDomain(@Valid @RequestBody Domain domain){
        Domain createdDomain  = sharingRegistryService.createDomain(domain);
        return new ResponseEntity<>(createdDomain, HttpStatus.CREATED);
    }

    /**
     <p>API method to update a domain</p>
     */
    @RequestMapping(method = RequestMethod.PUT)
    public ResponseEntity<String> updateDomain(@Valid @RequestBody Domain domain){
        sharingRegistryService.updateDomain(domain);
        return new ResponseEntity<>(HttpStatus.OK);
    }

    /**
     <p>API method to check Domain Exists</p>
     */
    @RequestMapping(value = "/{domainId}", method = RequestMethod.HEAD)
    public ResponseEntity<String> isDomainExists(@PathVariable("domainId") String domainId) {
        boolean response = sharingRegistryService.isDomainExists(domainId);
        HttpStatus status = HttpStatus.NOT_FOUND;
        if(response){
            status = HttpStatus.OK;
        }
        return new ResponseEntity<>(status);
    }

    /**
     <p>API method to retrieve a domain</p>
     */
    @RequestMapping(value = "/{domainId}",method = RequestMethod.GET)
    public ResponseEntity<Domain> getDomain(@PathVariable("domainId") String domainId){
        Domain domain = sharingRegistryService.getDomain(domainId);
        return new ResponseEntity<>(domain, HttpStatus.OK);
    }

    /**
     <p>API method to delete domain</p>
     */
    @RequestMapping(value = "/{domainId}", method = RequestMethod.DELETE)
    public ResponseEntity<String> deleteDomain(@PathVariable("domainId") String domainId){
        sharingRegistryService.deleteDomain(domainId);
        return new ResponseEntity<>(HttpStatus.OK);
    }

    /**
     <p>API method to get all domains.</p>
     */
    @RequestMapping(method = RequestMethod.GET)
    public @ResponseBody List<Domain> getDomains(@RequestParam(name = "offset", defaultValue = "0") int offset, @RequestParam(name = "limit", defaultValue = "-1") int limit){
        return sharingRegistryService.getDomains(offset, limit);
    }

}

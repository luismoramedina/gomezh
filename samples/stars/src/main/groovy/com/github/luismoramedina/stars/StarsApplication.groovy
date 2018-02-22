package com.github.luismoramedina.stars

import groovy.util.logging.Slf4j
import org.springframework.boot.SpringApplication
import org.springframework.boot.autoconfigure.SpringBootApplication
import org.springframework.web.bind.annotation.PathVariable
import org.springframework.web.bind.annotation.RequestHeader
import org.springframework.web.bind.annotation.RequestMapping
import org.springframework.web.bind.annotation.RequestMethod
import org.springframework.web.bind.annotation.RestController

@RestController
@SpringBootApplication
@Slf4j
class StarsApplication {

	static void main(String[] args) {
		SpringApplication.run StarsApplication, args
	}


	@RequestMapping(value = "/stars/{id}", method=RequestMethod.GET)
	Star star(
			@PathVariable String id,
			@RequestHeader(value="Authorization", required = false) String auth,
			@RequestHeader(value="X-B3-TraceId", required = false) String traceId,
			@RequestHeader(value="Plain-authorization", required = false) String plainAuth) {

		log.info "traceId: " + traceId
		log.info "plainAuth: " + plainAuth
		log.info "auth: " + auth
		def star = new Star()
		star.number = 5
		star.id = id
		star
	}
}

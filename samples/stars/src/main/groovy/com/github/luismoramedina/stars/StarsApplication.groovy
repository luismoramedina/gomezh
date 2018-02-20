package com.github.luismoramedina.stars

import org.springframework.boot.SpringApplication
import org.springframework.boot.autoconfigure.SpringBootApplication
import org.springframework.web.bind.annotation.RequestMapping
import org.springframework.web.bind.annotation.RestController

@RestController
@SpringBootApplication
class StarsApplication {

	static void main(String[] args) {
		SpringApplication.run StarsApplication, args
	}

	@RequestMapping
	Star star() {
		def star = new Star()
		star.number = 5
		star.id = 1
		star
	}
}

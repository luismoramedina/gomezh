package com.github.luismoramedina.meshless;

import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.context.annotation.Bean;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.client.RestTemplate;

import java.util.ArrayList;
import java.util.List;

@SpringBootApplication
@RestController
@RequestMapping("/books")
@Slf4j
public class MeshlessBooksApplication {

    @Autowired
    RestTemplate restTemplate;

	@Value("${stars.service.uri}")
	private String url;

	public static void main(String[] args) {
		SpringApplication.run(MeshlessBooksApplication.class, args);
	}

	@Bean
	public RestTemplate restTemplate() {
		return new RestTemplate();
	}

    @RequestMapping(method = RequestMethod.GET)
	public List<Book> books() {
		log.info("Before calling " + url);
		Star stars = restTemplate.getForObject(url, Star.class, 1);

		ArrayList<Book> books = new ArrayList<>();
		Book endersGame = new Book();
		endersGame.id = 1;
		endersGame.author = "orson scott card";
		endersGame.title = "Enders game";
		endersGame.year = "1985";
		endersGame.stars = stars.number;
		books.add(endersGame);
		return books;
	}

    @RequestMapping(method = RequestMethod.POST, consumes = "application/json")
    @ResponseBody
    @ResponseStatus(HttpStatus.CREATED)
    public Book newBook(@RequestBody Book book) {
        log.info("new book: " + book);
        return book;
    }
}
